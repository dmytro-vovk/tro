package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Config struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
	SSLMode  string
}

func New(config Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.SSLMode,
	))
	if err != nil {
		return nil, fmt.Errorf("error conecting to database: %s", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %s", err)
	}

	return db, nil
}
