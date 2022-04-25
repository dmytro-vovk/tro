package webserver

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"log"
	"net/http"
	"time"
)

type Auth interface {
	Valid(username string, password string) bool
}

type Webserver struct {
	server *http.Server
	config *viper.Viper
	logger *logrus.Logger
	writer io.WriteCloser
}

type Option interface {
	apply(*Webserver)
}

type optionFunc func(*Webserver)

func (fn optionFunc) apply(w *Webserver) {
	fn(w)
}

func New(addr string, handler http.Handler, logger *logrus.Logger, options ...Option) *Webserver {
	writer := logger.WriterLevel(logrus.ErrorLevel)

	server := &Webserver{
		server: &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  30 * time.Second,
			Addr:         addr,
			Handler:      handler,
			ErrorLog:     log.New(writer, "", 0),
		},
		logger: logger,
		writer: writer,
	}

	for _, opt := range options {
		opt.apply(server)
	}

	return server
}

func WithTLS(handler http.Handler, v *viper.Viper) Option {
	var config *tls.Config
	if v.GetBool("tls.enabled") {
		manager := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(v.GetString("tls.cert_dir")),
			Email:      "dmytro.vovk@pm.me",
			HostPolicy: autocert.HostWhitelist("tro.pw"),
		}

		handler = manager.HTTPHandler(handler)
		config = manager.TLSConfig()
	}

	return optionFunc(func(w *Webserver) {
		w.server.Handler = handler
		w.server.TLSConfig = config
	})
}

func (w *Webserver) Serve(name string) (err error) {
	w.logger.Infof("%s starting at %s", name, w.server.Addr)

	started := make(chan struct{})

	if w.server.TLSConfig != nil {
		w.server.Addr = ":443"
		go func() {
			if err := w.server.ListenAndServeTLS("", ""); err != nil {
				// todo: ошибка такого рода должна пробрасываться в main
				w.logger.Fatalf("error running TLS: %s", err)
			}
			close(started)
		}()
	} else {
		close(started)
	}

	<-started
	w.server.Addr = ":80"
	err = w.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	return
}

func (w *Webserver) Stop(ctx context.Context) error {
	if err := w.writer.Close(); err != nil {
		return err
	}

	return w.server.Shutdown(ctx)
}
