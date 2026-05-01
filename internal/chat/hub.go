package chat

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type     string      `json:"type"`
	Message  interface{} `json:"message,omitempty"`
	Messages interface{} `json:"messages,omitempty"`
	Content  string      `json:"content,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type IncomingMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type Client struct {
	Conn    *websocket.Conn
	UserID  uuid.UUID
	GroupID uuid.UUID
	Send    chan []byte
}

type Hub struct {
	rooms      map[uuid.UUID]map[*Client]bool
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan BroadcastMsg
}

type BroadcastMsg struct {
	GroupID uuid.UUID
	Data    []byte
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]map[*Client]bool),
		Register:   make(chan *Client, 64),
		Unregister: make(chan *Client, 64),
		Broadcast:  make(chan BroadcastMsg, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.rooms[client.GroupID] == nil {
				h.rooms[client.GroupID] = make(map[*Client]bool)
			}
			h.rooms[client.GroupID][client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.GroupID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)
				}
				if len(room) == 0 {
					delete(h.rooms, client.GroupID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.rooms[msg.GroupID] {
				select {
				case client.Send <- msg.Data:
				default:
					close(client.Send)
					delete(h.rooms[msg.GroupID], client)
				}
			}
			h.mu.RUnlock()
		}
	}
}
