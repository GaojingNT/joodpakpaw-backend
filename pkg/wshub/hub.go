package wshub

import (
	"fmt"
	"sync"

	"github.com/gofiber/contrib/websocket"
	socketio "github.com/zishang520/socket.io/v2/socket"
)

var AppHub *Hub

func CurrentHub() *Hub {
	return AppHub
}

type Hub struct {
	clients    map[int]*websocket.Conn
	sseClients map[int]chan []byte
	sioServer  *socketio.Server
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int]*websocket.Conn),
		sseClients: make(map[int]chan []byte),
	}
}

func (h *Hub) SetSocketIOServer(s *socketio.Server) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sioServer = s
}

func (h *Hub) GetSocketIOServer() *socketio.Server {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sioServer
}

func (h *Hub) Register(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = conn
}

func (h *Hub) Unregister(userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

func (h *Hub) RegisterSSE(userID int) chan []byte {
	ch := make(chan []byte, 10)
	h.mu.Lock()
	if oldCh, ok := h.sseClients[userID]; ok {
		close(oldCh)
	}
	h.sseClients[userID] = ch
	h.mu.Unlock()
	return ch
}

func (h *Hub) UnregisterSSE(userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.sseClients[userID]; ok {
		close(ch)
		delete(h.sseClients, userID)
	}
}

func (h *Hub) SendToUser(userID int, message []byte) {
	h.mu.RLock()
	wsConn, wsOk := h.clients[userID]
	sseCh, sseOk := h.sseClients[userID]
	sio := h.sioServer
	h.mu.RUnlock()

	if wsOk {
		wsConn.WriteMessage(websocket.TextMessage, message)
	}

	if sseOk {
		select {
		case sseCh <- message:
		default:
		}
	}

	if sio != nil {
		room := socketio.Room(fmt.Sprintf("user_%d", userID))
		sio.To(room).Emit("notification", string(message))
	}
}

func (h *Hub) EmitToUser(userID int, event string, data interface{}) {
	h.mu.RLock()
	sio := h.sioServer
	h.mu.RUnlock()

	if sio != nil {
		room := socketio.Room(fmt.Sprintf("user_%d", userID))
		sio.To(room).Emit(event, data)
	}
}
