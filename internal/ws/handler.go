package ws

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func ServeWS(hub *Hub, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("ws: upgrade failed", slog.String("error", err.Error()))
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		client.hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}
