package boot

import (
	"context"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"net/http"
	"os"
	"sync"
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
)

var boot *Boot

type Boot struct {
	*Container
	Shutdown func()
}

func New() *Boot {
	if boot != nil {
		return boot
	}

	c, fn := NewContainer()
	boot = &Boot{Container: c, Shutdown: fn}

	*logrus.StandardLogger() = *boot.Logger()
	logrus.RegisterExitHandler(fn)

	return boot
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

func (c *Boot) API(log *logrus.Logger) *api.Handler {
	const id = "API"
	if s, ok := c.Get(id).(*api.Handler); ok {
		return s
	}

	handler := api.NewHandler(log, c.APIService())

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

	l := c.Logger()
	w := l.WriterLevel(logrus.ErrorLevel)
	s := webserver.New(
		c.Config().WebServer.Listen,
		c.WebRouter(),
		w,
	)

	c.Set(id, s, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Stop(ctx); err != nil {
			logrus.Println("Error stopping web server:", err)
		}

		if err := w.Close(); err != nil {
			logrus.Println("Error closing web server error logger:", err)
		}
	})

	return s
}

func (c *Boot) APIRouter(log *logrus.Logger) http.Handler {
	const id = "API Router"
	if s, ok := c.Get(id).(http.Handler); ok {
		return s
	}

	r := c.API(log).Router()
	c.Set(id, r, nil)

	return r
}

func (c *Boot) APIServer() *webserver.Webserver {
	const id = "API Server"
	if s, ok := c.Get(id).(*webserver.Webserver); ok {
		return s
	}

	l := c.Logger()
	w := l.WriterLevel(logrus.ErrorLevel)
	s := webserver.New(
		c.Config().API.Listen,
		c.APIRouter(l),
		w,
	)

	c.Set(id, s, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Stop(ctx); err != nil {
			logrus.Println("Error stopping API server:", err)
		}

		if err := w.Close(); err != nil {
			logrus.Println("Error closing API server error logger:", err)
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
		logrus.Fatalf("Error connecting to database: %s", err)
	}

	c.Set(id, db, func() {
		if err := db.Close(); err != nil {
			logrus.Printf("Error closing database: %s", err)
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
		logrus.Fatalf("Error creating repository: %s", err)
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
		logrus.Fatalf("Error creating API service: %s", err)
	}

	c.Set(id, s, nil)

	return s
}

var logger struct {
	*logrus.Logger
	once sync.Once
}

func (c *Boot) Logger() *logrus.Logger {
	logger.once.Do(func() {
		const id = "Logger"

		var (
			path            = "/var/log/tro/tro.log"
			timestampFormat = "2006-01-02 15:04:05"
			fieldMap        = logrus.FieldMap{
				logrus.FieldKeyMsg: "message",
			}
		)

		logger.Logger = &logrus.Logger{
			Out: io.Discard,
			Formatter: &runtime.Formatter{
				ChildFormatter: &logrus.TextFormatter{
					ForceColors:     true,
					FullTimestamp:   true,
					TimestampFormat: timestampFormat,
					FieldMap:        fieldMap,
				},
				Line:         true,
				Package:      true,
				File:         true,
				BaseNameOnly: true,
			},
			Hooks:    logrus.LevelHooks{},
			Level:    logrus.DebugLevel,
			ExitFunc: os.Exit,
		}

		rotor := &lumberjack.Logger{
			Filename:   path,
			MaxSize:    1,
			MaxAge:     1,
			MaxBackups: 3,
			Compress:   true,
		}

		for _, hook := range []logrus.Hook{
			// Send logs with level higher than warning to stderr
			&writer.Hook{
				Writer: os.Stderr,
				LogLevels: []logrus.Level{
					logrus.PanicLevel,
					logrus.FatalLevel,
					logrus.ErrorLevel,
					logrus.WarnLevel,
				},
			},
			// Send info and debug logs to stdout
			&writer.Hook{
				Writer: os.Stdout,
				LogLevels: []logrus.Level{
					logrus.InfoLevel,
					logrus.DebugLevel,
				},
			},
			// Send all logs to file in JSON format with rotation
			lfshook.NewHook(rotor, &runtime.Formatter{
				ChildFormatter: &logrus.JSONFormatter{
					TimestampFormat: timestampFormat,
					FieldMap:        fieldMap,
				},
				Line:         true,
				Package:      true,
				File:         true,
				BaseNameOnly: true,
			}),
		} {
			logger.AddHook(hook)
		}

		c.Set(id, rotor, func() {
			if err := rotor.Rotate(); err != nil {
				logrus.Println("Error rotating log files:", err)
			}

			if err := rotor.Close(); err != nil {
				logrus.Println("Error closing log files rotator:", err)
			}
		})
	})

	return logger.Logger
}
