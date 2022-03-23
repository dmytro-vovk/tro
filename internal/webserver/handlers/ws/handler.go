package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/michcald/go-tools/internal/webserver/handlers/ws/client"
)

type Handler struct {
	client *client.Client
}

func NewHandler(c *client.Client) *Handler {
	return &Handler{client: c}
}

// Handler handles the websockets
func (h *Handler) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := (&websocket.Upgrader{
		EnableCompression: true,
		CheckOrigin:       func(*http.Request) bool { return true },
	}).Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection to websocket: %s", err)
		return
	}

	h.client.Run(conn)
}