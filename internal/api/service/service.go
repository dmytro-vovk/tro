package service

import (
	"fmt"

	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/dmytro-vovk/tro/internal/api/repository"
	v1 "github.com/dmytro-vovk/tro/internal/api/service/v1"
	v2 "github.com/dmytro-vovk/tro/internal/api/service/v2"
)

type Authorization interface {
	CreateUser(user model.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type Service interface {
	Authorization
}

func New(db repository.Repository, authMethod string) (Service, error) {
	switch authMethod {
	case "jwt":
		return v1.New(db), nil
	case "avigilon":
		return v2.New(db), nil
	default:
		return nil, fmt.Errorf("authorization method %q doesn't exist", authMethod)
	}
}
