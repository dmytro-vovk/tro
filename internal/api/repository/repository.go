package repository

import (
	"fmt"
	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/dmytro-vovk/tro/internal/api/repository/mysql"
	"github.com/dmytro-vovk/tro/internal/api/repository/postgres"
	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user model.User) (int, error)
	GetUser(username, password string) (model.User, error)
}

type Repository interface {
	Authorization
}

func New(db *sqlx.DB) (Repository, error) {
	switch db.DriverName() {
	case "mysql":
		return mysql.New(db), nil
	case "postgres":
		return postgres.New(db), nil
	default:
		return nil, fmt.Errorf("repository %q driver isn't exist", db.DriverName())
	}
}
