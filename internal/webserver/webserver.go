package webserver

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"time"
)

type Auth interface {
	Valid(username string, password string) bool
}

type Webserver struct {
	listen string
	server *http.Server
}

func New(listen string, handler http.Handler, tls *tls.Config, writer io.Writer) *Webserver {
	return &Webserver{
		listen: listen,
		server: &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  30 * time.Second,
			Addr:         listen,
			Handler:      handler,
			ErrorLog:     log.New(writer, "", 0),
			TLSConfig:    tls,
		},
	}
}

func (w *Webserver) Serve(name string) (err error) {
	logrus.Printf("%s starting at %s", name, w.listen)

	started := make(chan struct{})

	if w.server.TLSConfig != nil {
		w.server.Addr = ":443"
		go func() {
			if err := w.server.ListenAndServeTLS("", ""); err != nil {
				log.Fatalf("Error running TLS: %s", err)
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

func (w *Webserver) Stop(ctx context.Context) error { return w.server.Shutdown(ctx) }
