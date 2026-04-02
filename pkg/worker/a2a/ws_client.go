package a2a

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// ConnectionState represents the client's connection lifecycle.
type ConnectionState string

const (
	StateConnected    ConnectionState = "connected"
	StateConnecting   ConnectionState = "connecting"
	StateDisconnected ConnectionState = "disconnected"
)

// StateRecoverer is implemented by components that need to recover
// in-flight task state after a reconnection (e.g., the Server).
type StateRecoverer interface {
	// InFlightTasks returns tasks with status "working" or "input-required".
	InFlightTasks() []Task
	// OnStateRecovered is called after reconnect with the recovered task list.
	OnStateRecovered(tasks []Task)
}

// ClientConfig holds configuration for the A2A WebSocket client.
type ClientConfig struct {
	Transport      TransportConfig
	AgentCard      AgentCard
	StateRecoverer StateRecoverer
}

// Client wraps Transport with automatic reconnection and state recovery.
type Client struct {
	transport *Transport
	card      AgentCard
	recoverer StateRecoverer

	state ConnectionState
	mu    sync.Mutex

	onConnected       func()
	onDisconnected    func(error)
	onReconnectFailed func(error)
}

// NewClient creates a new A2A WebSocket client.
func NewClient(config ClientConfig) *Client {
	return &Client{
		transport: NewTransport(config.Transport),
		card:      config.AgentCard,
		recoverer: config.StateRecoverer,
		state:     StateDisconnected,
	}
}

// OnConnected registers a callback fired on initial connect and each reconnect.
func (c *Client) OnConnected(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onConnected = fn
}

// OnDisconnected registers a callback fired when the connection is lost.
func (c *Client) OnDisconnected(fn func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onDisconnected = fn
}

// OnReconnectFailed registers a callback fired when all retry attempts are exhausted.
func (c *Client) OnReconnectFailed(fn func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onReconnectFailed = fn
}

// State returns the current connection state.
func (c *Client) State() ConnectionState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// Transport returns the underlying transport for direct message I/O.
func (c *Client) Transport() *Transport {
	return c.transport
}

// Connect establishes the initial WebSocket connection, registers the
// agent card, and transitions to the connected state.
func (c *Client) Connect(ctx context.Context) error {
	c.setState(StateConnecting)

	if err := c.transport.Connect(ctx); err != nil {
		c.setState(StateDisconnected)
		return fmt.Errorf("client connect: %w", err)
	}

	if err := c.registerCard(); err != nil {
		c.setState(StateDisconnected)
		return fmt.Errorf("client register card: %w", err)
	}

	c.setState(StateConnected)
	c.fireConnected()
	return nil
}

// HandleConnectionLoss should be called when a receive or send error
// indicates the connection is broken. It triggers automatic reconnection
// with state recovery. Returns nil on successful reconnect.
func (c *Client) HandleConnectionLoss(ctx context.Context, cause error) error {
	c.setState(StateDisconnected)
	c.fireDisconnected(cause)

	log.Printf("[a2a] connection lost: %v — starting reconnection", cause)

	return c.reconnectAndRecover(ctx)
}

// Close gracefully shuts down the client and its transport.
func (c *Client) Close() error {
	c.setState(StateDisconnected)
	return c.transport.Close()
}

// reconnectAndRecover attempts to reconnect using the transport's
// exponential backoff, then re-registers the agent card and recovers
// in-flight task state.
func (c *Client) reconnectAndRecover(ctx context.Context) error {
	c.setState(StateConnecting)

	if err := c.transport.Reconnect(ctx); err != nil {
		c.setState(StateDisconnected)
		c.fireReconnectFailed(err)
		return fmt.Errorf("client reconnect: %w", err)
	}

	if err := c.registerCard(); err != nil {
		c.setState(StateDisconnected)
		c.fireReconnectFailed(err)
		return fmt.Errorf("client re-register card: %w", err)
	}

	c.recoverState()
	c.setState(StateConnected)
	c.fireConnected()

	log.Printf("[a2a] reconnected and state recovered")
	return nil
}

// registerCard sends the agent card registration over the transport.
func (c *Client) registerCard() error {
	params, err := marshalJSON(c.card)
	if err != nil {
		return fmt.Errorf("marshal agent card: %w", err)
	}
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  MethodRegisterCard,
		Params:  params,
	}
	data, err := marshalJSON(req)
	if err != nil {
		return fmt.Errorf("marshal register request: %w", err)
	}
	return c.transport.Send(data)
}

// recoverState collects in-flight tasks from the StateRecoverer and
// reports their status back to the backend after reconnection.
func (c *Client) recoverState() {
	if c.recoverer == nil {
		return
	}

	tasks := c.recoverer.InFlightTasks()
	if len(tasks) == 0 {
		log.Printf("[a2a] no in-flight tasks to recover")
		return
	}

	log.Printf("[a2a] recovering %d in-flight task(s)", len(tasks))

	for _, t := range tasks {
		params := StatusUpdateParams{
			TaskID: t.ID,
			Status: t.Status,
		}
		notif := JSONRPCNotification{
			JSONRPC: "2.0",
			Method:  MethodStatusUpdate,
			Params:  params,
		}
		data, err := marshalJSON(notif)
		if err != nil {
			log.Printf("[a2a] marshal recovery update for %s: %v", t.ID, err)
			continue
		}
		if err := c.transport.Send(data); err != nil {
			log.Printf("[a2a] send recovery update for %s: %v", t.ID, err)
		}
	}

	c.recoverer.OnStateRecovered(tasks)
}

// setState updates the connection state under the mutex.
func (c *Client) setState(s ConnectionState) {
	c.mu.Lock()
	c.state = s
	c.mu.Unlock()
}

// fireConnected invokes the OnConnected callback if set.
func (c *Client) fireConnected() {
	c.mu.Lock()
	fn := c.onConnected
	c.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// fireDisconnected invokes the OnDisconnected callback if set.
func (c *Client) fireDisconnected(err error) {
	c.mu.Lock()
	fn := c.onDisconnected
	c.mu.Unlock()
	if fn != nil {
		fn(err)
	}
}

// fireReconnectFailed invokes the OnReconnectFailed callback if set.
func (c *Client) fireReconnectFailed(err error) {
	c.mu.Lock()
	fn := c.onReconnectFailed
	c.mu.Unlock()
	if fn != nil {
		fn(err)
	}
}
