package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}

func New(config Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	))
	if err != nil {
		return nil, fmt.Errorf("error conecting to database: %s", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %s", err)
	}

	return db, nil
}
