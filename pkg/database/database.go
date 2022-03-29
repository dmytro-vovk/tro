package database

import (
	"fmt"
	"github.com/dmytro-vovk/tro/pkg/database/mysql"
	"github.com/dmytro-vovk/tro/pkg/database/postgres"
	"github.com/jmoiron/sqlx"
	"os"
)

func New(driver string) (*sqlx.DB, error) {
	switch driver {
	case "mysql":
		return mysql.New(mysql.Config{
			Host:     os.Getenv("MYSQL_HOST"),
			Port:     os.Getenv("MYSQL_PORT"),
			Username: os.Getenv("MYSQL_USERNAME"),
			Password: os.Getenv("MYSQL_PASSWORD"),
			Database: os.Getenv("MYSQL_DATABASE"),
		})
	case "postgres":
		return postgres.New(postgres.Config{
			Host:     os.Getenv("POSTGRESQL_HOST"),
			Port:     os.Getenv("POSTGRESQL_PORT"),
			Username: os.Getenv("POSTGRESQL_USERNAME"),
			Password: os.Getenv("POSTGRESQL_PASSWORD"),
			Database: os.Getenv("POSTGRESQL_DATABASE"),
			SSLMode:  os.Getenv("POSTGRESQL_SSLMODE"),
		})
	default:
		return nil, fmt.Errorf("unknown database driver name %q", driver)
	}
}
