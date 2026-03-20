package adapter

import (
	"context"
	"fmt"
	"sync"
)

// Registry는 PlatformAdapter 레지스트리이다.
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]PlatformAdapter
}

// NewRegistry는 빈 레지스트리를 생성한다.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]PlatformAdapter),
	}
}

// Register는 어댑터를 등록한다.
func (r *Registry) Register(a PlatformAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[a.Name()] = a
}

// Get은 이름으로 어댑터를 조회한다.
func (r *Registry) Get(name string) (PlatformAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[name]
	if !ok {
		return nil, fmt.Errorf("adapter %q not found", name)
	}
	return a, nil
}

// List는 등록된 모든 어댑터를 반환한다.
func (r *Registry) List() []PlatformAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]PlatformAdapter, 0, len(r.adapters))
	for _, a := range r.adapters {
		result = append(result, a)
	}
	return result
}

// DetectAll은 설치된 코딩 CLI를 감지하여 해당 어댑터만 반환한다.
func (r *Registry) DetectAll(ctx context.Context) []PlatformAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var detected []PlatformAdapter
	for _, a := range r.adapters {
		if ok, err := a.Detect(ctx); err == nil && ok {
			detected = append(detected, a)
		}
	}
	return detected
}
