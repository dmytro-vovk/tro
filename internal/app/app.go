package app

import (
	"github.com/dmytro-vovk/tro/internal/storage"
)

type Application struct {
	responseCounter int
	streamer        Streamer
	storage         *storage.Storage
}

type Streamer interface {
	Notify(string, interface{})
}

func New(db *storage.Storage) *Application {
	return &Application{
		storage: db,
	}
}

func (a *Application) SetStreamer(s Streamer) {
	a.streamer = s
	go a.pinger()
}
