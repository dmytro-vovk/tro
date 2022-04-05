package webserver

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
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

func New(listen string, handler http.Handler, writer io.Writer) *Webserver {
	return &Webserver{
		listen: listen,
		server: &http.Server{
			Addr:     listen,
			Handler:  handler,
			ErrorLog: log.New(writer, "", 0),
		},
	}
}

func (w *Webserver) Serve(name string) (err error) {
	logrus.Printf("%s starting at %s", name, w.listen)

	err = w.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	return
}

func (w *Webserver) Stop(ctx context.Context) error { return w.server.Shutdown(ctx) }
