package lore

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// QueryContext는 특정 파일 경로와 관련된 Lore 항목을 반환한다.
func QueryContext(dir, path string) ([]LoreEntry, error) {
	if err := checkGitDir(dir); err != nil {
		return nil, err
	}

	output, err := runGitLog(dir, "--", path)
	if err != nil {
		return nil, err
	}

	entries := parseGitLogOutput(output)
	return entries, nil
}

// QueryConstraints는 Constraint 트레일러가 있는 Lore 항목을 반환한다.
func QueryConstraints(dir string) ([]LoreEntry, error) {
	return queryByField(dir, "Constraint")
}

// QueryRejected는 Rejected 트레일러가 있는 Lore 항목을 반환한다.
func QueryRejected(dir string) ([]LoreEntry, error) {
	return queryByField(dir, "Rejected")
}

// QueryDirectives는 Directive 트레일러가 있는 Lore 항목을 반환한다.
func QueryDirectives(dir string) ([]LoreEntry, error) {
	return queryByField(dir, "Directive")
}

// QueryStale은 N일보다 오래된 Lore 항목을 반환한다.
func QueryStale(dir string, days int) ([]LoreEntry, error) {
	if err := checkGitDir(dir); err != nil {
		return nil, err
	}

	output, err := runGitLog(dir)
	if err != nil {
		return nil, err
	}

	all := parseGitLogOutput(output)
	cutoff := time.Now().AddDate(0, 0, -days)

	stale := make([]LoreEntry, 0)
	for _, e := range all {
		if !e.CommitDate.IsZero() && e.CommitDate.Before(cutoff) {
			stale = append(stale, e)
		}
	}
	return stale, nil
}

// queryByField는 특정 트레일러 필드가 있는 항목을 필터링한다.
func queryByField(dir, fieldKey string) ([]LoreEntry, error) {
	if err := checkGitDir(dir); err != nil {
		return nil, err
	}

	output, err := runGitLog(dir)
	if err != nil {
		return nil, err
	}

	all := parseGitLogOutput(output)

	var result []LoreEntry
	for _, e := range all {
		if hasField(e, fieldKey) {
			result = append(result, e)
		}
	}
	return result, nil
}

// hasField는 LoreEntry에 특정 필드 값이 있는지 확인한다.
func hasField(e LoreEntry, fieldKey string) bool {
	switch fieldKey {
	case "Constraint":
		return e.Constraint != ""
	case "Rejected":
		return e.Rejected != ""
	case "Confidence":
		return e.Confidence != ""
	case "Directive":
		return e.Directive != ""
	case "Tested":
		return e.Tested != ""
	}
	return false
}

// runGitLog는 git log 명령을 실행하고 출력을 반환한다.
func runGitLog(dir string, extraArgs ...string) (string, error) {
	// %H=해시, %ai=날짜(ISO), %B=전체 메시지
	args := []string{"log", "--format=COMMIT:%H%n DATE:%ai%n%B%nEND_COMMIT"}
	args = append(args, extraArgs...)

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git log 실행 실패: %w\n%s", err, stderr.String())
	}
	return out.String(), nil
}

// parseGitLogOutput는 git log 출력을 LoreEntry 목록으로 파싱한다.
func parseGitLogOutput(output string) []LoreEntry {
	var entries []LoreEntry

	// COMMIT: 구분자로 분리
	commits := strings.Split(output, "COMMIT:")
	for _, block := range commits {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		if len(lines) < 2 {
			continue
		}

		hash := strings.TrimSpace(lines[0])
		var dateStr, msgBody string

		// DATE: 라인 찾기
		msgLines := lines[1:]
		for i, l := range msgLines {
			if strings.HasPrefix(l, " DATE:") {
				dateStr = strings.TrimSpace(strings.TrimPrefix(l, " DATE:"))
				msgLines = msgLines[i+1:]
				break
			}
		}

		// END_COMMIT 이전까지 메시지 수집
		var cleanLines []string
		for _, l := range msgLines {
			if l == "END_COMMIT" {
				break
			}
			cleanLines = append(cleanLines, l)
		}
		msgBody = strings.Join(cleanLines, "\n")

		entry, err := ParseTrailers(msgBody)
		if err != nil {
			continue
		}
		entry.CommitHash = hash
		entry.CommitMsg = extractSubject(msgBody)

		if dateStr != "" {
			entry.CommitDate, _ = time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		}

		entries = append(entries, *entry)
	}
	return entries
}

// extractSubject는 커밋 메시지에서 첫 번째 줄(제목)을 추출한다.
func extractSubject(msg string) string {
	lines := strings.Split(strings.TrimSpace(msg), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

// checkGitDir는 디렉터리가 git 저장소인지 확인한다.
func checkGitDir(dir string) error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git 저장소가 아닙니다: %s", dir)
	}
	return nil
}
