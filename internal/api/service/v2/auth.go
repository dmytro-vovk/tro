package v2

import (
	"github.com/dmytro-vovk/tro/internal/api/model"
)

func (s *service) CreateUser(user model.User) (int, error) {
	return 0, nil
}

func (s *service) GenerateToken(username, password string) (string, error) {
	return "", nil
}

func (s *service) ParseToken(accessToken string) (int, error) {
	return 0, nil
}
