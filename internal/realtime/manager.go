// Package realtime provides WebSocket-based real-time subscriptions for Fieldstone.
// It allows clients to subscribe to collection changes and receive instant updates.
package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Event represents a database change event
type Event struct {
	Type         string      `json:"type"`         // create, update, delete
	CollectionID string      `json:"collectionId"`
	RecordID     string      `json:"recordId"`
	Data         interface{} `json:"data,omitempty"`
	Timestamp    int64       `json:"timestamp"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	Rooms    map[string]bool // collection IDs the client is subscribed to
	mu       sync.RWMutex
	manager  *Manager
}

// Manager handles WebSocket connections and message broadcasting
type Manager struct {
	clients    map[string]*Client
	rooms      map[string]map[string]*Client // room (collectionID) -> clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	mu         sync.RWMutex
	upgrader   websocket.Upgrader
}

// NewManager creates a new realtime manager
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event, 256),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in development
				// In production, check against allowed origins
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// Run starts the manager's event loop
func (m *Manager) Run() {
	log.Info().Msg("Realtime manager started")
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.ID] = client
			m.mu.Unlock()
			log.Debug().Str("client", client.ID).Msg("Client registered")

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client.ID]; ok {
				// Remove from all rooms
				for room := range client.Rooms {
					delete(m.rooms[room], client.ID)
					if len(m.rooms[room]) == 0 {
						delete(m.rooms, room)
					}
				}
				delete(m.clients, client.ID)
				close(client.Send)
			}
			m.mu.Unlock()
			log.Debug().Str("client", client.ID).Msg("Client unregistered")

		case event := <-m.broadcast:
			m.broadcastToRoom(event)
		}
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket")
		return
	}

	client := &Client{
		ID:      generateClientID(),
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Rooms:   make(map[string]bool),
		manager: m,
	}

	m.register <- client

	// Start goroutines for reading and writing
	go client.readPump()
	go client.writePump()
}

// readPump handles incoming messages from the client
func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Str("client", c.ID).Msg("WebSocket error")
			}
			break
		}

		// Handle incoming message
		c.handleMessage(message)
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Message types
const (
	MessageTypeSubscribe   = "subscribe"
	MessageTypeUnsubscribe = "unsubscribe"
	MessageTypeEvent       = "event"
	MessageTypeError       = "error"
	MessageTypeAck         = "ack"
)

// IncomingMessage represents a message from the client
type IncomingMessage struct {
	Type         string `json:"type"`
	CollectionID string `json:"collectionId"`
}

// OutgoingMessage represents a message to the client
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (c *Client) handleMessage(data []byte) {
	var msg IncomingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("Invalid message format")
		return
	}

	switch msg.Type {
	case MessageTypeSubscribe:
		c.subscribe(msg.CollectionID)
	case MessageTypeUnsubscribe:
		c.unsubscribe(msg.CollectionID)
	default:
		c.sendError("Unknown message type: " + msg.Type)
	}
}

func (c *Client) subscribe(collectionID string) {
	c.mu.Lock()
	c.Rooms[collectionID] = true
	c.mu.Unlock()

	c.manager.mu.Lock()
	if c.manager.rooms[collectionID] == nil {
		c.manager.rooms[collectionID] = make(map[string]*Client)
	}
	c.manager.rooms[collectionID][c.ID] = c
	c.manager.mu.Unlock()

	log.Debug().
		Str("client", c.ID).
		Str("collection", collectionID).
		Msg("Client subscribed to collection")

	// Send acknowledgment
	ack := OutgoingMessage{
		Type: MessageTypeAck,
		Payload: map[string]string{
			"action":       "subscribed",
			"collectionId": collectionID,
		},
	}
	data, _ := json.Marshal(ack)
	c.Send <- data
}

func (c *Client) unsubscribe(collectionID string) {
	c.mu.Lock()
	delete(c.Rooms, collectionID)
	c.mu.Unlock()

	c.manager.mu.Lock()
	if room, ok := c.manager.rooms[collectionID]; ok {
		delete(room, c.ID)
		if len(room) == 0 {
			delete(c.manager.rooms, collectionID)
		}
	}
	c.manager.mu.Unlock()

	log.Debug().
		Str("client", c.ID).
		Str("collection", collectionID).
		Msg("Client unsubscribed from collection")

	// Send acknowledgment
	ack := OutgoingMessage{
		Type: MessageTypeAck,
		Payload: map[string]string{
			"action":       "unsubscribed",
			"collectionId": collectionID,
		},
	}
	data, _ := json.Marshal(ack)
	c.Send <- data
}

func (c *Client) sendError(message string) {
	err := OutgoingMessage{
		Type: MessageTypeError,
		Payload: map[string]string{
			"message": message,
		},
	}
	data, _ := json.Marshal(err)
	c.Send <- data
}

// Broadcast sends an event to all subscribed clients
func (m *Manager) Broadcast(event Event) {
	select {
	case m.broadcast <- event:
	default:
		log.Warn().Msg("Broadcast channel full, dropping event")
	}
}

func (m *Manager) broadcastToRoom(event Event) {
	m.mu.RLock()
	room, ok := m.rooms[event.CollectionID]
	m.mu.RUnlock()

	if !ok {
		return
	}

	msg := OutgoingMessage{
		Type:    MessageTypeEvent,
		Payload: event,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal event")
		return
	}

	for _, client := range room {
		select {
		case client.Send <- data:
		default:
			// Client send buffer full, close connection
			close(client.Send)
			m.unregister <- client
		}
	}
}

// GetStats returns realtime system statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"clients":    len(m.clients),
		"rooms":      len(m.rooms),
		"broadcasts": len(m.broadcast),
	}
}

func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

// Subscription represents a database subscription for the backend interface
type Subscription struct {
	CollectionID string
	Cancel       context.CancelFunc
}

// Unsubscribe cancels the subscription
func (s *Subscription) Unsubscribe() error {
	if s.Cancel != nil {
		s.Cancel()
	}
	return nil
}
