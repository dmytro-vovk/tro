package webserver

import (
	"context"
	"errors"
	"log"
	"net/http"
)

type Auth interface {
	Valid(username string, password string) bool
}

type Webserver struct {
	listen string
	server *http.Server
}

func New(handler http.Handler, listen string) *Webserver {
	return &Webserver{
		listen: listen,
		server: &http.Server{
			Addr:    listen,
			Handler: handler,
		},
	}
}

func (w *Webserver) Serve(name string) (err error) {
	log.Printf("%s starting at %s", name, w.listen)

	err = w.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	return
}

func (w *Webserver) Stop(ctx context.Context) error { return w.server.Shutdown(ctx) }
