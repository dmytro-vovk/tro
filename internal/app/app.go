package app

import (
	"github.com/jmoiron/sqlx"
)

type Application struct {
	responseCounter int
	streamer        Streamer
	db              *sqlx.DB
}

type Streamer interface {
	Notify(string, interface{})
}

func New(db *sqlx.DB) *Application {
	return &Application{
		db: db,
	}
}

func (a *Application) SetStreamer(s Streamer) {
	a.streamer = s
	go a.pinger()
}
