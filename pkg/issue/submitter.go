package issue

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
)

// CommandRunner abstracts os/exec for testability.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
	LookPath(name string) (string, error)
}

// defaultRunner is the real implementation using os/exec.
type defaultRunner struct{}

func (d *defaultRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func (d *defaultRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// Submitter orchestrates GitHub issue creation via the gh CLI.
type Submitter struct {
	runner CommandRunner
}

// NewSubmitter creates a Submitter with the given CommandRunner.
// Pass nil to use the real os/exec runner.
func NewSubmitter(runner CommandRunner) *Submitter {
	if runner == nil {
		runner = &defaultRunner{}
	}
	return &Submitter{runner: runner}
}

// CheckGH verifies that gh is installed and authenticated.
func (s *Submitter) CheckGH() error {
	path, err := s.runner.LookPath("gh")
	if err != nil || path == "" {
		return fmt.Errorf("gh CLI is not installed or not in PATH: %w", err)
	}

	out, err := s.runner.Run("gh", "auth", "status")
	if err != nil {
		return fmt.Errorf("gh CLI is not authenticated — run `gh auth login`: %w", err)
	}
	_ = out
	return nil
}

// ComputeHash returns the xxhash hex digest of (errMsg + command).
func (s *Submitter) ComputeHash(errMsg, cmd string) string {
	h := xxhash.Sum64String(errMsg + "\x00" + cmd)
	return fmt.Sprintf("%016x", h)
}

// ghIssueItem is used to unmarshal `gh issue list` JSON output.
type ghIssueItem struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

// FindDuplicate searches for an existing issue with the given hash label.
// Returns the issue URL if found, or empty string if not found.
func (s *Submitter) FindDuplicate(repo, hash string) (string, error) {
	out, err := s.runner.Run("gh", "issue", "list",
		"--repo", repo,
		"--label", "auto-report",
		"--search", hash,
		"--json", "number,url",
		"--limit", "1",
	)
	if err != nil {
		return "", fmt.Errorf("find duplicate: gh issue list: %w", err)
	}

	var issues []ghIssueItem
	if err := json.Unmarshal(out, &issues); err != nil {
		return "", fmt.Errorf("find duplicate: parse response: %w", err)
	}

	if len(issues) == 0 {
		return "", nil
	}
	return issues[0].URL, nil
}

// CreateIssue creates a new GitHub issue and returns the result.
func (s *Submitter) CreateIssue(repo, title, body string, labels []string) (SubmitResult, error) {
	args := []string{"issue", "create",
		"--repo", repo,
		"--title", title,
		"--body", body,
	}
	for _, l := range labels {
		args = append(args, "--label", l)
	}

	out, err := s.runner.Run("gh", args...)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("create issue: %w", err)
	}

	issueURL := strings.TrimSpace(string(out))
	num := parseIssueNumber(issueURL)

	return SubmitResult{
		IssueURL:    issueURL,
		IssueNumber: num,
		Action:      "created",
	}, nil
}

// AddComment adds a comment to an existing issue.
func (s *Submitter) AddComment(repo string, issueNum int, body string) error {
	_, err := s.runner.Run("gh", "issue", "comment",
		strconv.Itoa(issueNum),
		"--repo", repo,
		"--body", body,
	)
	if err != nil {
		return fmt.Errorf("add comment to issue #%d: %w", issueNum, err)
	}
	return nil
}

// Submit orchestrates CheckGH → hash → FindDuplicate → CreateIssue or AddComment.
func (s *Submitter) Submit(report IssueReport, body string) (SubmitResult, error) {
	if err := s.CheckGH(); err != nil {
		return SubmitResult{}, err
	}

	dupURL, err := s.FindDuplicate(report.Repo, report.Hash)
	if err != nil {
		return SubmitResult{}, err
	}

	if dupURL != "" {
		num := parseIssueNumber(dupURL)
		if err := s.AddComment(report.Repo, num, body); err != nil {
			return SubmitResult{}, err
		}
		return SubmitResult{
			IssueURL:     dupURL,
			IssueNumber:  num,
			WasDuplicate: true,
			Action:       "commented",
		}, nil
	}

	return s.CreateIssue(report.Repo, report.Title, body, report.Labels)
}

// parseIssueNumber extracts the issue number from a GitHub issue URL.
func parseIssueNumber(url string) int {
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
	if len(parts) == 0 {
		return 0
	}
	n, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0
	}
	return n
}
