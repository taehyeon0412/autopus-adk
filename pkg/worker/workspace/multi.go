package workspace

import (
	"fmt"
	"sync"
)

// WorkspaceConn represents a connection to a single workspace.
type WorkspaceConn struct {
	WorkspaceID string
	ProjectDir  string
	BackendURL  string
	AuthToken   string
	Connected   bool
}

// MultiWorkspace manages connections to multiple workspaces.
// All methods are safe for concurrent use.
type MultiWorkspace struct {
	conns map[string]*WorkspaceConn
	mu    sync.RWMutex
}

// NewMultiWorkspace creates an empty multi-workspace manager.
func NewMultiWorkspace() *MultiWorkspace {
	return &MultiWorkspace{
		conns: make(map[string]*WorkspaceConn),
	}
}

// Add registers a workspace connection.
// If a connection with the same WorkspaceID already exists, it is replaced.
func (mw *MultiWorkspace) Add(conn WorkspaceConn) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.conns[conn.WorkspaceID] = &conn
}

// Get returns the connection for a workspace ID.
// Returns false if not found.
func (mw *MultiWorkspace) Get(workspaceID string) (*WorkspaceConn, bool) {
	mw.mu.RLock()
	defer mw.mu.RUnlock()
	conn, ok := mw.conns[workspaceID]
	return conn, ok
}

// RouteTask returns the project directory for a workspace-targeted task.
// Returns an error if the workspace is not registered or not connected.
func (mw *MultiWorkspace) RouteTask(workspaceID string) (string, error) {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	conn, ok := mw.conns[workspaceID]
	if !ok {
		return "", fmt.Errorf("workspace %q not registered", workspaceID)
	}
	if !conn.Connected {
		return "", fmt.Errorf("workspace %q not connected", workspaceID)
	}
	return conn.ProjectDir, nil
}

// List returns all registered workspace connections.
// The returned slice is a snapshot; mutations do not affect the manager.
func (mw *MultiWorkspace) List() []WorkspaceConn {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	result := make([]WorkspaceConn, 0, len(mw.conns))
	for _, conn := range mw.conns {
		result = append(result, *conn)
	}
	return result
}

// Remove removes a workspace connection.
// No-op if the workspace ID is not registered.
func (mw *MultiWorkspace) Remove(workspaceID string) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	delete(mw.conns, workspaceID)
}
