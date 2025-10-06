package ws

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageEvent string

type WSConnectionManager struct {
	Connections map[uuid.UUID]*websocket.Conn
	mu          sync.RWMutex
}

func NewWSConnectionManager() *WSConnectionManager {
	return &WSConnectionManager{
		Connections: make(map[uuid.UUID]*websocket.Conn),
	}
}

// AddConnection adds a new user connection to the manager
func (wm *WSConnectionManager) AddConnection(userId uuid.UUID, conn *websocket.Conn) {
	wm.mu.Lock()
	wm.Connections[userId] = conn
	wm.mu.Unlock()
}

// RemoveAndCloseConnection removes a user connection from the manager and closes it
func (wm *WSConnectionManager) RemoveAndCloseConnection(userId uuid.UUID) {
	wm.mu.Lock()
	if _, exists := wm.Connections[userId]; exists {
		delete(wm.Connections, userId)
	}
	wm.mu.Unlock()
}

func (wm *WSConnectionManager) IsConnected(userId uuid.UUID) (*websocket.Conn, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	conn, exists := wm.Connections[userId]
	return conn, exists
}
