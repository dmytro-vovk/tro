package boot

import (
	"context"
	"log"
	"time"

	"github.com/michcald/go-tools/internal/app"
	"github.com/michcald/go-tools/internal/webserver"
	"github.com/michcald/go-tools/internal/webserver/handlers/ws"
	"github.com/michcald/go-tools/internal/webserver/handlers/ws/client"
)

type Boot struct {
	*Container
}

func New() (*Boot, func()) {
	c, fn := NewContainer()
	boot := Boot{
		Container: c,
	}

	return &boot, fn
}

func (c *Boot) Application() *app.Application {
	id := "Application"
	if s, ok := c.Get(id).(*app.Application); ok {
		return s
	}

	a := app.New()

	c.Set(id, a, nil)

	return a
}

func (c *Boot) Webserver() *webserver.Webserver {
	id := "Web Server"
	if s, ok := c.Get(id).(*webserver.Webserver); ok {
		return s
	}

	s := webserver.New(
		c.WebsocketHandler(),
		":8080",
	)

	fn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.Stop(ctx); err != nil {
			log.Printf("Error stopping web server: %s", err)
		}
	}

	c.Set(id, s, fn)

	return s
}

func (c *Boot) WebsocketHandler() *ws.Handler {
	id := "WS Handler"
	if s, ok := c.Get(id).(*ws.Handler); ok {
		return s
	}

	h := ws.NewHandler(c.WSClient())

	c.Set(id, h, nil)

	return h
}

func (c *Boot) WSClient() *client.Client {
	id := "WS Client"
	if s, ok := c.Get(id).(*client.Client); ok {
		return s
	}

	s := client.New().
		NS("example",
			client.NSMethod("method", c.Application().Example),
		)

	c.Application().SetStreamer(s)

	c.Set(id, s, nil)

	return s
}
