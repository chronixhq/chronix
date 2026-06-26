package agentmux

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"sync"
	"time"

	"runtime/debug"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Manager keeps track of live agent connections by ID (UUID) and stores display name as metadata.
// Duplicate IDs will replace the older connection with the newer one.
// All logs use standard slog keys: component, op, id, agent, error.

// Manager maintains the active websocket connections for all connected agents.
type Manager struct {
	mu      sync.RWMutex
	agents  map[string]*Conn // key: uuid
	closed  bool
	ctx     context.Context
	cancel  context.CancelFunc
	pingInt time.Duration
}

// OnConnect is an optional callback invoked when an agent connection is registered.
// The function receives (uuid, name). Set via SetOnConnect.
var OnConnect func(string, string)

// OnDisconnect is an optional callback invoked when an agent connection is unregistered.
// The function receives (uuid, name). Set via SetOnDisconnect.
var OnDisconnect func(string, string)

var DefaultManager = NewManager()

// NewManager creates and initializes a new agent Manager.
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		agents:  make(map[string]*Conn),
		ctx:     ctx,
		cancel:  cancel,
		pingInt: 30 * time.Second,
	}
}

// SetPingInterval allows overriding the server ping interval.
func (m *Manager) SetPingInterval(d time.Duration) { m.pingInt = d } // test 2

// Register adds a new websocket connection to the manager, replacing any existing connection for the same ID.
func (m *Manager) Register(id string, name string, ws *websocket.Conn, remoteAddr net.Addr) *Conn {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		_ = ws.Close()
		return nil
	}
	if old := m.agents[id]; old != nil {
		// Replace old connection
		slog.Info("agent replace", "component", "agentmux", "op", "replace", "id", id, "agent", name)
		old.Close()
	}
	conn := newConn(m.ctx, id, name, ws, m.pingInt)
	m.agents[id] = conn
	slog.Info("agent connected", "component", "agentmux", "op", "register", "id", id, "agent", name, "remote", remoteAddr.String())
	if OnConnect != nil {
		go OnConnect(id, name)
	}
	conn.start(func() { m.onConnClosed(id, conn) })
	return conn
}

func (m *Manager) onConnClosed(id string, c *Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cur, ok := m.agents[id]; ok && cur == c {
		delete(m.agents, id)
		slog.Info("agent disconnected", "component", "agentmux", "op", "unregister", "id", id, "agent", c.name)
		if OnDisconnect != nil {
			name := c.name
			go OnDisconnect(id, name)
		}
	}
}

func (m *Manager) Unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c := m.agents[id]; c != nil {
		c.Close()
		delete(m.agents, id)
	}
}

func (m *Manager) Get(id string) *Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.agents[id]
}

func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := make([]string, 0, len(m.agents))
	for k := range m.agents {
		n = append(n, k)
	}
	return n
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.agents)
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true
	m.cancel()
	for id, c := range m.agents {
		slog.Info("closing agent", "component", "agentmux", "op", "shutdown", "id", id, "agent", c.name)
		c.Close()
	}
	m.agents = map[string]*Conn{}
	m.mu.Unlock()
}

// Conn wraps a single agent websocket connection.

type Conn struct {
	id       string
	name     string
	ws       *websocket.Conn
	send     chan []byte
	ctx      context.Context
	cancel   context.CancelFunc
	pingInt  time.Duration
	lastSeen time.Time

	// request/response correlation
	wmu     sync.Mutex
	waiters map[string]chan respMsg
}

// SendAndWait is a thin wrapper over Request for convenience in typed flows.
func (c *Conn) SendAndWait(ctx context.Context, msgType string, payload any) (string, []byte, error) {
	return c.Request(ctx, msgType, payload)
}

// NotifyAgentDeleted sends a deletion request to the agent and waits for an ACK.
// Returns (acked, wasOnline, err). If the agent is not online, (false, false, nil) is returned.
func (m *Manager) NotifyAgentDeleted(id, reason string, timeout time.Duration) (bool, bool, error) {
	m.mu.RLock()
	c := m.agents[id]
	m.mu.RUnlock()
	if c == nil {
		return false, false, nil
	}
	payload := struct {
		Reason string `json:"reason"`
		When   string `json:"when"`
	}{Reason: reason, When: time.Now().Format(time.RFC3339)}
	ctx, cancel := context.WithTimeout(m.ctx, timeout)
	defer cancel()
	t, _, err := c.SendAndWait(ctx, "agent.deleted", payload)
	if err != nil {
		slog.Warn("notify delete error", "component", "agentmux", "op", "notify-delete", "id", id, "error", err)
		return false, true, err
	}
	if t == "agent.deleted.ack" {
		return true, true, nil
	}
	return false, true, nil
}

func newConn(parent context.Context, id string, name string, ws *websocket.Conn, pingInt time.Duration) *Conn {
	ctx, cancel := context.WithCancel(parent)
	return &Conn{
		id:      id,
		name:    name,
		ws:      ws,
		send:    make(chan []byte, 128),
		ctx:     ctx,
		cancel:  cancel,
		pingInt: pingInt,
		waiters: make(map[string]chan respMsg),
	}
}

func (c *Conn) start(onClose func()) {
	go c.readLoop(onClose)
	go c.writeLoop()
	go c.pingLoop()
}

type respMsg struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

type reqEnvelope struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Payload any    `json:"payload"`
}

func (c *Conn) readLoop(onClose func()) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recovered from panic in agent readLoop", "agent_id", c.id, "agent_name", c.name, "error", r, "stack", string(debug.Stack()))
		}
		_ = c.ws.Close()
		onClose()
	}()
	c.ws.SetReadLimit(1 << 20) // 1 MiB
	_ = c.ws.SetReadDeadline(time.Now().Add(c.pingInt * 3))
	c.ws.SetPongHandler(func(string) error {
		c.lastSeen = time.Now()
		return c.ws.SetReadDeadline(time.Now().Add(c.pingInt * 3))
	})
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			t, data, err := c.ws.ReadMessage()
			if err != nil {
				return
			}
			if t != websocket.TextMessage {
				c.lastSeen = time.Now()
				continue
			}
			var rm respMsg
			if err := json.Unmarshal(data, &rm); err == nil && rm.ID != "" {
				c.wmu.Lock()
				ch := c.waiters[rm.ID]
				if ch != nil {
					// non-blocking send
					select {
					case ch <- rm:
					default:
					}
				}
				c.wmu.Unlock()
			}
			c.lastSeen = time.Now()
		}
	}
}

func (c *Conn) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recovered from panic in agent writeLoop", "agent_id", c.id, "agent_name", c.name, "error", r, "stack", string(debug.Stack()))
		}
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			_ = c.ws.SetWriteDeadline(time.Now().Add(15 * time.Second))
			if err := c.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

// Request sends a message to the agent and waits for a response with a timeout.
func (c *Conn) Request(ctx context.Context, msgType string, payload any) (string, []byte, error) {
	id := uuid.NewString()
	respCh := make(chan respMsg, 1)
	c.wmu.Lock()
	c.waiters[id] = respCh
	c.wmu.Unlock()
	defer func() {
		c.wmu.Lock()
		delete(c.waiters, id)
		c.wmu.Unlock()
		close(respCh)
	}()
	req := reqEnvelope{Type: msgType, ID: id, Payload: payload}
	data, err := json.Marshal(req)
	if err != nil {
		return "", nil, err
	}
	select {
	case c.send <- data:
		// sent
	case <-ctx.Done():
		return "", nil, ctx.Err()
	}
	select {
	case rm := <-respCh:
		return rm.Type, rm.Payload, nil
	case <-ctx.Done():
		// best-effort cancel message
		cancel := reqEnvelope{Type: "sql.cancel", ID: id, Payload: nil}
		if b, _ := json.Marshal(cancel); b != nil {
			select {
			case c.send <- b:
			default:
			}
		}
		return "", nil, ctx.Err()
	}
}

func (c *Conn) pingLoop() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recovered from panic in agent pingLoop", "agent_id", c.id, "agent_name", c.name, "error", r, "stack", string(debug.Stack()))
		}
	}()
	t := time.NewTicker(c.pingInt)
	defer t.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-t.C:
			_ = c.ws.SetWriteDeadline(time.Now().Add(15 * time.Second))
			if err := c.ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}

func (c *Conn) ID() string {
	return c.id
}

func (c *Conn) Close() {
	c.cancel()
	_ = c.ws.Close()
	close(c.send)
}
