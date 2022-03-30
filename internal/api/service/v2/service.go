package v2

import "github.com/dmytro-vovk/tro/internal/api/repository"

type service struct {
	db repository.Repository
}

func New(db repository.Repository) *service {
	return &service{db: db}
}
