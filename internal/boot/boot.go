package boot

import (
	"context"
	"fmt"
	"github.com/dmytro-vovk/tro/internal/boot/config"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmytro-vovk/tro/internal/api"
	"github.com/dmytro-vovk/tro/internal/api/repository"
	"github.com/dmytro-vovk/tro/internal/api/service"
	"github.com/dmytro-vovk/tro/internal/app"
	"github.com/dmytro-vovk/tro/internal/webserver"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/home"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/ws"
	"github.com/dmytro-vovk/tro/internal/webserver/handlers/ws/client"
	"github.com/dmytro-vovk/tro/internal/webserver/router"
	"github.com/spf13/viper"
)

type boot struct {
	container
	viper  *viper.Viper
	logger *logrus.Logger
}

func New() (*boot, error) {
	b := &boot{viper: config.DefaultViper()}

	if err := b.loadEnv(); err != nil {
		return nil, err
	}

	if err := b.loadConfig(); err != nil {
		return nil, err
	}

	if err := b.configureLogger(); err != nil {
		return nil, err
	}

	go b.shutdown()

	return b, nil
}

func (b *boot) shutdown() {
	logrus.RegisterExitHandler(b.shutdown)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	logrus.Infof("Got %v, shutting down...", s)
	b.container.shutdown()
	os.Exit(0)
}

func (b *boot) loadEnv() error {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(".env")
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error loading env variables: %w", err)
	}

	for key := range v.AllSettings() {
		if err := b.viper.BindEnv(key); err != nil {
			return fmt.Errorf("error binding env variable %q: %w", key, err)
		}
	}

	if err := b.viper.MergeConfigMap(v.AllSettings()); err != nil {
		return fmt.Errorf("error merging with env config: %w", err)
	}

	return nil
}

func (b *boot) loadConfig() error {
	v := viper.New()
	v.AddConfigPath("configs")

	in := "dev-config"
	if s := b.viper.GetString("config"); s != "" {
		in = s
	}
	v.SetConfigName(in)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error loading %q config: %w", in, err)
	}

	if err := b.viper.MergeConfigMap(v.AllSettings()); err != nil {
		return fmt.Errorf("error merging with %q config: %w", in, err)
	}

	return nil
}

func (b *boot) Application() *app.Application {
	const id = "Application"
	if s, ok := b.Get(id).(*app.Application); ok {
		return s
	}

	a := app.New(nil)

	b.Set(id, a, nil)

	return a
}

func (b *boot) WebRouter() http.Handler {
	const id = "Web Router"
	if s, ok := b.Get(id).(http.Handler); ok {
		return s
	}

	r := router.New(
		router.Route("/ws", b.WebsocketHandler().Handler),
		router.Route("/js/index.js", home.Scripts),
		router.Route("/js/index.js.map", home.ScriptsMap),
		router.Route("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		router.RoutePrefix("/css/", home.Styles.ServeHTTP),
		router.CatchAll(home.Handler),
	)

	b.Set(id, r, nil)

	return r
}

func (b *boot) Webserver() (*webserver.Webserver, error) {
	const id = "Web Server"
	if s, ok := b.Get(id).(*webserver.Webserver); ok {
		return s, nil
	}

	handler := b.WebRouter()
	server := webserver.New(b.viper.GetString("webserver.listen"), handler, b.logger, webserver.WithTLS(handler, b.viper))

	b.Set(id, server, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logrus.Println("Error stopping web server:", err)
		}
	})

	return server, nil
}

func (b *boot) APIServer() (*webserver.Webserver, error) {
	const id = "API Server"
	if server, ok := b.Get(id).(*webserver.Webserver); ok {
		return server, nil
	}

	repo, err := repository.New(b.viper.Sub("database"))
	if err != nil {
		return nil, fmt.Errorf("can't create API repository: %w", err)
	}

	srv, err := service.New(repo, b.viper.Sub("api"))
	if err != nil {
		return nil, fmt.Errorf("can't create API service: %w", err)
	}

	server := webserver.New(b.viper.GetString("api.listen"), api.NewHandler(b.logger, srv).Router(), b.logger)

	b.Set(id, server, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := repo.Close(); err != nil {
			b.logger.Errorf("error closing %s database: %s", repo.DriverName(), err)
		}

		if err := server.Stop(ctx); err != nil {
			b.logger.Errorln("error stopping API server:", err)
		}
	})

	return server, nil
}

func (b *boot) WebsocketHandler() *ws.Handler {
	const id = "WS Handler"
	if s, ok := b.Get(id).(*ws.Handler); ok {
		return s
	}

	h := ws.NewHandler(b.WSClient())

	b.Set(id, h, nil)

	return h
}

func (b *boot) WSClient() *client.Client {
	const id = "WS Client"
	if s, ok := b.Get(id).(*client.Client); ok {
		return s
	}

	s := client.New().
		NS("example",
			client.NSMethod("method", b.Application().Example),
		).
		NS("code",
			client.NSMethod("generate_image", b.Application().QR),
		)

	b.Application().SetStreamer(s)

	b.Set(id, s, nil)

	return s
}

func (b *boot) configureLogger() error {
	var cfg config.Logger
	if err := b.viper.UnmarshalKey("logger", &cfg); err != nil {
		return fmt.Errorf("unable to unmarshall logger config: %w", err)
	}

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("can't parse logger level: %w", err)
	}

	logger := &logrus.Logger{
		Level: level,
		Out:   io.Discard,
		Formatter: cfg.Formatter(&logrus.TextFormatter{
			ForceColors:     true,
			DisableQuote:    true,
			FullTimestamp:   true,
			TimestampFormat: cfg.TimestampFormat,
			FieldMap:        cfg.FieldMap(),
		}),
		Hooks:    make(logrus.LevelHooks),
		ExitFunc: os.Exit,
	}

	// Configure logs rotation
	rotor := &lumberjack.Logger{
		Filename:   cfg.Rotor.Filename,
		MaxSize:    cfg.Rotor.MaxSize,
		MaxAge:     cfg.Rotor.MaxAge,
		MaxBackups: cfg.Rotor.MaxBackups,
		LocalTime:  cfg.Rotor.LocalTime,
		Compress:   cfg.Rotor.Compress,
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
		lfshook.NewHook(rotor, cfg.Formatter(&logrus.JSONFormatter{
			TimestampFormat: cfg.TimestampFormat,
			FieldMap:        cfg.FieldMap(),
		})),
	} {
		logger.AddHook(hook)
	}

	b.Set("logger", rotor, func() {
		if err := rotor.Rotate(); err != nil {
			b.logger.Errorln("error rotating log files:", err)
		}

		if err := rotor.Close(); err != nil {
			b.logger.Errorln("error closing log files rotator:", err)
		}
	})

	b.logger = logger
	b.logger.WithFields(logrus.Fields{
		"output": rotor.Filename,
		"grade":  b.logger.GetLevel(),
	}).Info("logger was successfully configured")

	return nil
}
