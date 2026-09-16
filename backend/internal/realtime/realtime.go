// Package realtime delivers job events to connected technician clients over
// WebSocket. It implements notifications.Service so the jobs service can
// publish events without knowing anything about transport.
package realtime

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jobman/backend/internal/notifications"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 1 << 12
	sendBuffer     = 32
)

// Client is a single connected device.
type Client struct {
	conn    *websocket.Conn
	techID  string
	send    chan []byte
	onClose func(techID string, c *Client)
	once    sync.Once
}

// EnqueuedMessage is the wire envelope pushed to clients.
type EnqueuedMessage struct {
	Type string         `json:"type"`
	Time int64          `json:"ts"`
	Job  notifications.JobEvent `json:"job"`
}

// Hub tracks live connections per technician and fans out messages.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*Client]struct{})}
}

func (h *Hub) register(techID string, c *Client) {
	c.onClose = func(t string, cl *Client) { h.unregister(t, cl) }
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[techID] == nil {
		h.clients[techID] = make(map[*Client]struct{})
	}
	h.clients[techID][c] = struct{}{}
}

func (h *Hub) unregister(techID string, c *Client) {
	c.once.Do(func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if set := h.clients[techID]; set != nil {
			delete(set, c)
			if len(set) == 0 {
				delete(h.clients, techID)
			}
		}
		close(c.send)
	})
}

// Broadcaster routes a message to every socket of a technician.
type Broadcaster interface {
	BroadcastToTechnician(techID string, msg EnqueuedMessage)
}

func (h *Hub) BroadcastToTechnician(techID string, msg EnqueuedMessage) {
	if techID == "" {
		return
	}
	payload, err := jsonMarshal(msg)
	if err != nil {
		log.Printf("realtime: marshal: %v", err)
		return
	}
	h.mu.RLock()
	set := h.clients[techID]
	targets := make([]*Client, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		select {
		case c.send <- payload:
		default:
			// Slow client: drop it rather than blocking the broadcaster.
			log.Printf("realtime: dropping slow client %s", techID)
			go c.close()
		}
	}
}

func (c *Client) close() {
	c.once.Do(func() {
		if c.onClose != nil {
			c.onClose(c.techID, c)
		}
		_ = c.conn.Close()
	})
}

// writePump sends queued messages and periodic pings.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()
	for {
		select {
		case payload, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump detects disconnects and drains inbound frames.
func (c *Client) readPump() {
	defer c.close()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *Client) start() {
	go c.writePump()
	go c.readPump()
}

// NotifyJobAssigned implements notifications.Service.
func (h *Hub) NotifyJobAssigned(_ context.Context, e notifications.JobEvent) {
	h.BroadcastToTechnician(e.TechnicianID, EnqueuedMessage{
		Type: "job_assigned",
		Time: time.Now().Unix(),
		Job:  e,
	})
}

// NotifyJobStatusChanged implements notifications.Service.
func (h *Hub) NotifyJobStatusChanged(_ context.Context, e notifications.JobEvent) {
	h.BroadcastToTechnician(e.TechnicianID, EnqueuedMessage{
		Type: "job_status_changed",
		Time: time.Now().Unix(),
		Job:  e,
	})
}

// NotifyJobCompleted implements notifications.Service.
func (h *Hub) NotifyJobCompleted(_ context.Context, e notifications.JobEvent) {
	h.BroadcastToTechnician(e.TechnicianID, EnqueuedMessage{
		Type: "job_completed",
		Time: time.Now().Unix(),
		Job:  e,
	})
}

var _ notifications.Service = (*Hub)(nil)