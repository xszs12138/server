package livehub

import (
	"encoding/json"
	"sync"

	"blog-server/internal/dto"

	"github.com/gorilla/websocket"
)

// Hub 向所有已连接的博客客户端广播直播配置变更
type Hub struct {
	register   chan *client
	unregister chan *client
	broadcast  chan []byte

	mu      sync.RWMutex
	clients map[*client]struct{}
}

type client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func NewHub() *Hub {
	h := &Hub{
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan []byte, 16),
		clients:    make(map[*client]struct{}),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			targets := make([]*client, 0, len(h.clients))
			for c := range h.clients {
				targets = append(targets, c)
			}
			h.mu.RUnlock()

			for _, c := range targets {
				select {
				case c.send <- message:
				default:
					h.mu.Lock()
					if _, ok := h.clients[c]; ok {
						delete(h.clients, c)
						close(c.send)
						_ = c.conn.Close()
					}
					h.mu.Unlock()
				}
			}
		}
	}
}

// BroadcastLive 后台更新直播配置后推送给所有 WebSocket 客户端
func (h *Hub) BroadcastLive(item *dto.LiveBroadcast) {
	if h == nil || item == nil {
		return
	}
	payload, err := json.Marshal(dto.LiveWSMessage{
		Type: dto.LiveWSTypeUpdated,
		Data: item,
	})
	if err != nil {
		return
	}
	select {
	case h.broadcast <- payload:
	default:
	}
}

func (h *Hub) Register(conn *websocket.Conn) *client {
	c := &client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 8),
	}
	h.register <- c
	go c.writePump()
	return c
}

func (c *client) Unregister() {
	c.hub.unregister <- c
}

func (c *client) Enqueue(message []byte) {
	select {
	case c.send <- message:
	default:
	}
}

func (c *client) writePump() {
	defer func() {
		_ = c.conn.Close()
	}()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
