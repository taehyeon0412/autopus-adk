package worker

import (
	"os"
	"time"

	"github.com/insajin/autopus-adk/pkg/worker/security"
)

func (wl *WorkerLoop) taskExecutionDeadline(taskID string) (time.Time, bool) {
	timeout := wl.taskExecutionTimeout(taskID)
	if timeout <= 0 {
		return time.Time{}, false
	}

	cache := security.NewPolicyCache()
	info, err := os.Stat(cache.PolicyPath(taskID))
	if err != nil {
		return time.Now().Add(timeout), true
	}
	return info.ModTime().Add(timeout), true
}
