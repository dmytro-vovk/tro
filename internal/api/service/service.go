package service

import (
	"fmt"
	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/dmytro-vovk/tro/internal/api/repository"
	"github.com/dmytro-vovk/tro/internal/api/service/v1"
	"github.com/dmytro-vovk/tro/internal/api/service/v2"
)

type Authorization interface {
	CreateUser(user model.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type Service interface {
	Authorization
}

type Config struct {
	AuthMethod string
}

func New(db repository.Repository, config Config) (Service, error) {
	switch config.AuthMethod {
	case "jwt":
		return v1.New(db), nil
	case "avigilon":
		return v2.New(db), nil
	default:
		return nil, fmt.Errorf("authorization method %q isn't exist", config.AuthMethod)
	}
}
