package mysql

import (
	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/jmoiron/sqlx"
)

type storage struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *storage {
	return &storage{db: db}
}

func (s *storage) CreateUser(user model.User) (int, error) {
	return 0, nil
}

func (s *storage) GetUser(username, password string) (model.User, error) {
	return model.User{}, nil
}
