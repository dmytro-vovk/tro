package api

import (
	"log"
	"net/http"

	"github.com/dmytro-vovk/tro/internal/storage"
)

// API is for REST API server
type API struct {
	storage *storage.Storage
}

func New(db *storage.Storage) *API {
	return &API{
		storage: db,
	}
}

func (a *API) Handle404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.Log(r)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (a *API) Log(r *http.Request) {
	log.Printf("[%s] %s %s", r.RemoteAddr, r.Method, r.URL.Path)
}
