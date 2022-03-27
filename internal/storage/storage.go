package storage

import (
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	address string
	*sqlx.DB
}

func New(address string) *Storage {
	s := Storage{address: address}

	s.connect()

	return &s
}

func (s *Storage) connect() {
	var err error
	for {
		s.DB, err = sqlx.Open("mysql", s.address)
		if err != nil {
			log.Printf("Error conecting to DB: %s", err)
			time.Sleep(time.Second)

			continue
		}

		if err = s.DB.Ping(); err != nil {
			log.Printf("Error pinging DB: %s", err)
			time.Sleep(time.Second)

			continue
		}

		break
	}
}
