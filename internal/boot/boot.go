package boot

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dmytro-vovk/tro/internal/api"
	"github.com/dmytro-vovk/tro/internal/api/repository"
	"github.com/dmytro-vovk/tro/internal/api/service"
	"github.com/dmytro-vovk/tro/internal/app"
	"github.com/dmytro-vovk/tro/internal/boot/config"
	"github.com/dmytro-vovk/tro/internal/webserver"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/home"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/ws"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/ws/client"
	"github.com/dmytro-vovk/tro/internal/webserver/router"
	"github.com/dmytro-vovk/tro/pkg/database"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/acme/autocert"
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

func (c *Boot) WebRouter() http.Handler {
	const id = "Web Router"
	if s, ok := c.Get(id).(http.Handler); ok {
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

func (c *Boot) TLSManager() *autocert.Manager {
	const id = "TLS Manager"
	if s, ok := c.Get(id).(*autocert.Manager); ok {
		return s
	}

	tlsConfig := c.Config().WebServer.TLS

	s := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(tlsConfig.CertDir),
		Email:      "dmytro.vovk@pm.me",
		HostPolicy: autocert.HostWhitelist("tro.pw"),
	}

	c.Set(id, s, nil)

	return s
}

func (c *Boot) TLSConfig() *tls.Config {
	const id = "TLS Config"
	if s, ok := c.Get(id).(*tls.Config); ok {
		return s
	}

	tlsConfig := c.Config().WebServer.TLS

	if !tlsConfig.Enabled {
		return nil
	}

	s := c.TLSManager().TLSConfig()

	c.Set(id, s, nil)

	return s
}

func (c *Boot) Webserver() *webserver.Webserver {
	const id = "Web Server"
	if s, ok := c.Get(id).(*webserver.Webserver); ok {
		return s
	}

	handler := c.WebRouter()

	if c.Config().WebServer.TLS.Enabled {
		handler = c.TLSManager().HTTPHandler(handler)
	}

	s := webserver.New(
		handler,
		c.Config().WebServer.Listen,
		c.TLSConfig(),
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
		nil,
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
	const id = "Repository"
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
	const id = "API Service"
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
