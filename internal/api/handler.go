package api

import (
	"github.com/dmytro-vovk/tro/internal/api/handler/auth"
	"github.com/dmytro-vovk/tro/internal/api/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Authorization interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	UserIdentity(c *gin.Context)
}

// Handler is for REST Handler server
type Handler struct {
	auth Authorization
	log  *logrus.Logger
}

func NewHandler(log *logrus.Logger, serv service.Service) *Handler {
	return &Handler{
		auth: auth.NewHandler(log, serv),
		log:  log,
	}
}
