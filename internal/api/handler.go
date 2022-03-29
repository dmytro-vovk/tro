package api

import (
	"github.com/dmytro-vovk/tro/internal/api/handler/auth"
	"github.com/dmytro-vovk/tro/internal/api/service"
	"github.com/gin-gonic/gin"
)

type Authorization interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	UserIdentity(c *gin.Context)
}

// Handler is for REST Handler server
type Handler struct {
	auth Authorization
}

func NewHandler(serv service.Service) *Handler {
	return &Handler{auth: auth.NewHandler(serv)}
}
