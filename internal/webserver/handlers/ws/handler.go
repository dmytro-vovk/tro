package ws

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/dmytro-vovk/tro/internal/webserver/handlers/ws/client"
	"github.com/gorilla/websocket"
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
		logrus.Printf("Error upgrading connection to websocket: %s", err)
		return
	}

	h.client.Run(conn)
}
