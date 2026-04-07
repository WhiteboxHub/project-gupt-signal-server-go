package connection

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sampath/signaling-server/internal/models"
)

// Manager handles WebSocket connection lifecycle
type Manager struct {
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
}

// NewManager creates a new connection manager
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*websocket.Conn),
	}
}

// AddConnection registers a new WebSocket connection
func (m *Manager) AddConnection(connID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[connID] = conn
	log.Printf("[ConnectionManager] Added connection: %s (total: %d)", connID, len(m.connections))
}

// RemoveConnection unregisters a WebSocket connection
func (m *Manager) RemoveConnection(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, connID)
	log.Printf("[ConnectionManager] Removed connection: %s (total: %d)", connID, len(m.connections))
}

// GetConnection retrieves a connection by ID
func (m *Manager) GetConnection(connID string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[connID]
	return conn, exists
}

// SendMessage sends a message to a specific connection
func (m *Manager) SendMessage(connID string, msg *models.Message) error {
	conn, exists := m.GetConnection(connID)
	if !exists {
		return nil // Connection already closed
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// BroadcastToSession sends messages to both peers in a session
func (m *Manager) BroadcastToSession(hostConnID, clientConnID string, hostMsg, clientMsg *models.Message) {
	// Send to host
	if err := m.SendMessage(hostConnID, hostMsg); err != nil {
		log.Printf("[ConnectionManager] Failed to send to host %s: %v", hostConnID, err)
	}

	// Send to client
	if err := m.SendMessage(clientConnID, clientMsg); err != nil {
		log.Printf("[ConnectionManager] Failed to send to client %s: %v", clientConnID, err)
	}
}

// CloseConnection closes a WebSocket connection
func (m *Manager) CloseConnection(connID string) {
	conn, exists := m.GetConnection(connID)
	if exists {
		conn.Close()
		m.RemoveConnection(connID)
	}
}

// GetActiveConnections returns the count of active connections
func (m *Manager) GetActiveConnections() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}
