package boot

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dmytro-vovk/tro/internal/api"
	"github.com/dmytro-vovk/tro/internal/app"
	"github.com/dmytro-vovk/tro/internal/boot/config"
	"github.com/dmytro-vovk/tro/internal/storage"
	"github.com/dmytro-vovk/tro/internal/webserver"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/home"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/ws"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/ws/client"
	"github.com/dmytro-vovk/tro/internal/webserver/router"
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

func (c *Boot) Config() *config.Config {
	id := "Config"
	if s, ok := c.Get(id).(*config.Config); ok {
		return s
	}

	configFile := "config.json"
	if cfg := os.Getenv("CONFIG"); cfg != "" {
		configFile = cfg
	}

	s := config.Load(configFile)

	c.Set(id, s, nil)

	return s
}

func (c *Boot) Application() *app.Application {
	id := "Application"
	if s, ok := c.Get(id).(*app.Application); ok {
		return s
	}

	a := app.New(c.Storage())

	c.Set(id, a, nil)

	return a
}

func (c *Boot) API() *api.API {
	id := "API"
	if s, ok := c.Get(id).(*api.API); ok {
		return s
	}

	a := api.New(c.Storage())

	c.Set(id, a, nil)

	return a
}

func (c *Boot) WebHandler() http.HandlerFunc {
	id := "Web Router"
	if s, ok := c.Get(id).(http.HandlerFunc); ok {
		return s
	}

	r := router.New(
		router.Route("/ws", c.WebsocketHandler().Handler),
		router.Route("/js/index.js", home.Scripts),
		router.Route("/js/index.js.map", home.ScriptsMap),
		router.Route("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		router.RoutePrefix("/css/", home.Styles.ServeHTTP),
		router.CatchAll(home.Handler),
	)

	c.Set(id, r, nil)

	return r
}

func (c *Boot) Webserver() *webserver.Webserver {
	id := "Web Server"
	if s, ok := c.Get(id).(*webserver.Webserver); ok {
		return s
	}

	s := webserver.New(
		c.WebHandler(),
		c.Config().WebServer.Listen,
	)

	c.Set(id, s, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Stop(ctx); err != nil {
			log.Printf("Error stopping web server: %s", err)
		}
	})

	return s
}

func (c *Boot) APIHandler() http.HandlerFunc {
	id := "API Router"
	if s, ok := c.Get(id).(http.HandlerFunc); ok {
		return s
	}

	a := c.API()

	r := router.New(
		router.CatchAll(a.Handle404()),
	)

	c.Set(id, r, nil)

	return r
}

func (c *Boot) APIServer() *webserver.Webserver {
	id := "Web Server"
	if s, ok := c.Get(id).(*webserver.Webserver); ok {
		return s
	}

	s := webserver.New(
		c.APIHandler(),
		c.Config().API.Listen,
	)

	c.Set(id, s, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Stop(ctx); err != nil {
			log.Printf("Error stopping api server: %s", err)
		}
	})

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
		).
		NS("code",
			client.NSMethod("generate_image", c.Application().QR),
		)

	c.Application().SetStreamer(s)

	c.Set(id, s, nil)

	return s
}

func (c *Boot) Storage() *storage.Storage {
	id := "Database"
	if s, ok := c.Get(id).(*storage.Storage); ok {
		return s
	}

	s := storage.New(c.Config().Database.DSN)

	c.Set(id, s, func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing database: %s", err)
		}
	})

	return s
}
