package poll

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	httpTimeout   = 10 * time.Second
	maxBodyBytes  = 10 << 20 // 10MB
	emptyBodySize = 2        // "[]" or "{}" are effectively empty
)

// TaskPoller polls the backend REST API for pending tasks.
type TaskPoller struct {
	backendURL  string
	authToken   string
	workspaceID string
	backoff     *AdaptiveBackoff
	onTask      func(taskData []byte)
	client      *http.Client
}

// NewTaskPoller creates a poller that calls onTask when pending tasks are found.
func NewTaskPoller(backendURL, authToken, workspaceID string, onTask func([]byte)) *TaskPoller {
	return &TaskPoller{
		backendURL:  backendURL,
		authToken:   authToken,
		workspaceID: workspaceID,
		backoff:     NewAdaptiveBackoff(),
		onTask:      onTask,
		client:      &http.Client{Timeout: httpTimeout},
	}
}

// Start runs the polling loop until the context is cancelled.
func (p *TaskPoller) Start(ctx context.Context) {
	for {
		wait := p.backoff.Next()
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		data, err := p.fetchPending(ctx)
		if err != nil {
			log.Printf("[poll] fetch error: %v", err)
			continue
		}
		if len(data) <= emptyBodySize {
			continue
		}

		p.backoff.Reset()
		p.onTask(data)
	}
}

// Reset resets the backoff interval, typically called on a WebSocket push.
func (p *TaskPoller) Reset() {
	p.backoff.Reset()
}

func (p *TaskPoller) fetchPending(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/workspaces/%s/tasks/pending", p.backendURL, p.workspaceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.authToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxBodyBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}
