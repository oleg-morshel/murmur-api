package ws

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

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

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Хаб закрыл канал — отправляем close-фрейм и выходим
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.hub.log.Warn("ws: write failed", slog.String("error", err.Error()))
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.log.Warn("ws: ping failed", slog.String("error", err.Error()))
				return
			}
		}
	}
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	log        *logger.Logger
}

func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		log:        log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.log.Info("ws: client connected", slog.Int("total_clients", len(h.clients)))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.log.Info("ws: client disconnected", slog.Int("total_clients", len(h.clients)))
			}

		case msg := <-h.broadcast:
			h.log.Debug("ws: broadcasting",
				slog.Int("clients", len(h.clients)),
				slog.Int("bytes", len(msg)),
			)
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					delete(h.clients, client)
					close(client.send)
					h.log.Warn("ws: slow client dropped", slog.Int("total_clients", len(h.clients)))
				}
			}
		}
	}
}

func (h *Hub) Broadcast(data []byte) {
	h.broadcast <- data
}
