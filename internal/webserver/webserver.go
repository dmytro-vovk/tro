package webserver

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/michcald/go-tools/internal/webserver/handlers/home"
	"github.com/michcald/go-tools/internal/webserver/handlers/ws"
)

type Auth interface {
	Valid(username string, password string) bool
}

type Webserver struct {
	listen string
	server *http.Server
}

func New(handler *ws.Handler, listen string) *Webserver {
	return &Webserver{
		listen: listen,
		server: &http.Server{
			Addr: listen,
			Handler: NewRouter(
				Route("/ws", handler.Handler),
				Route("/js/index.js", home.Scripts),
				Route("/js/index.js.map", home.ScriptsMap),
				Route("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				}),
				RoutePrefix("/css/", home.Styles.ServeHTTP),
				CatchAll(home.Handler),
			),
		},
	}
}

func (w *Webserver) Serve() (err error) {
	log.Printf("Webserver starting at %s", w.listen)

	err = w.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	return
}

func (w *Webserver) Stop(ctx context.Context) error { return w.server.Shutdown(ctx) }
