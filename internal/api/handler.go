package api

import (
	"github.com/dmytro-vovk/tro/internal/api/handler/auth"
	"github.com/dmytro-vovk/tro/internal/api/repository"
	"github.com/dmytro-vovk/tro/internal/api/service"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Listen     string `json:"listen"`
	AuthMethod string `json:"auth_method"`
}

type Authorization interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	UserIdentity(c *gin.Context)
}

// Handler is for REST Handler server
type Handler struct {
	auth Authorization
}

func NewHandler(db *sqlx.DB, config Config) (*Handler, error) {
	repo, err := repository.New(db)
	if err != nil {
		return nil, err
	}

	serv, err := service.New(repo, service.Config{
		AuthMethod: config.AuthMethod,
	})
	if err != nil {
		return nil, err
	}

	return &Handler{auth: auth.NewHandler(serv)}, nil
}
