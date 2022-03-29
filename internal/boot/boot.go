package boot

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dmytro-vovk/tro/internal/api/repository"
	"github.com/dmytro-vovk/tro/internal/api/service"
	"github.com/dmytro-vovk/tro/pkg/database"
	"github.com/jmoiron/sqlx"

	"github.com/dmytro-vovk/tro/internal/api"
	"github.com/dmytro-vovk/tro/internal/app"
	"github.com/dmytro-vovk/tro/internal/boot/config"
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
	const id = "Config"
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
	const id = "Application"
	if s, ok := c.Get(id).(*app.Application); ok {
		return s
	}

	a := app.New(c.Storage())

	c.Set(id, a, nil)

	return a
}

func (c *Boot) API() *api.Handler {
	const id = "API"
	if s, ok := c.Get(id).(*api.Handler); ok {
		return s
	}

	handler := api.NewHandler(c.APIService())

	c.Set(id, handler, nil)

	return handler
}

func (c *Boot) WebRouter() http.HandlerFunc {
	const id = "Web Router"
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
	const id = "Web Server"
	if s, ok := c.Get(id).(*webserver.Webserver); ok {
		return s
	}

	s := webserver.New(
		c.WebRouter(),
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

func (c *Boot) APIRouter() http.Handler {
	const id = "API Router"
	if s, ok := c.Get(id).(http.Handler); ok {
		return s
	}

	r := c.API().Router()
	c.Set(id, r, nil)

	return r
}

func (c *Boot) APIServer() *webserver.Webserver {
	const id = "Web Server"
	if s, ok := c.Get(id).(*webserver.Webserver); ok {
		return s
	}

	s := webserver.New(
		c.APIRouter(),
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
	const id = "WS Handler"
	if s, ok := c.Get(id).(*ws.Handler); ok {
		return s
	}

	h := ws.NewHandler(c.WSClient())

	c.Set(id, h, nil)

	return h
}

func (c *Boot) WSClient() *client.Client {
	const id = "WS Client"
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

func (c *Boot) Storage() *sqlx.DB {
	const id = "Database"
	if s, ok := c.Get(id).(*sqlx.DB); ok {
		return s
	}

	db, err := database.New(c.Config().Database.DriverName)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	c.Set(id, db, func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %s", err)
		}
	})

	return db
}

func (c *Boot) Repository() repository.Repository {
	id := "Repository"
	if s, ok := c.Get(id).(repository.Repository); ok {
		return s
	}

	repo, err := repository.New(c.Storage())
	if err != nil {
		log.Fatalf("Error creating repository: %s", err)
	}

	c.Set(id, repo, nil)

	return repo
}

func (c *Boot) APIService() service.Service {
	id := "API Service"
	if s, ok := c.Get(id).(service.Service); ok {
		return s
	}

	s, err := service.New(c.Repository(), c.Config().API.AuthMethod)
	if err != nil {
		log.Fatalf("Error creating API service: %s", err)
	}

	c.Set(id, s, nil)

	return s
}
