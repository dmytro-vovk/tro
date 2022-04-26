package service

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"

	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/dmytro-vovk/tro/internal/api/repository"
	v1 "github.com/dmytro-vovk/tro/internal/api/service/v1"
	v2 "github.com/dmytro-vovk/tro/internal/api/service/v2"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go

type Authorization interface {
	CreateUser(user model.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type Service interface {
	Authorization
}

func New(db repository.Repository, v *viper.Viper) (Service, error) {
	if v == nil {
		return nil, errors.New("API configuration not provided")
	}

	switch method := v.GetString("auth_method"); method {
	case "jwt":
		return v1.New(db), nil
	case "avigilon":
		return v2.New(db), nil
	default:
		return nil, fmt.Errorf("authorization method %q doesn't exist", method)
	}
}
