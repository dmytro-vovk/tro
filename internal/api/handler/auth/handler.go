package auth

import (
	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/dmytro-vovk/tro/internal/api/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

// todo: make private
type Handler struct {
	auth service.Authorization
}

func NewHandler(serv service.Authorization) *Handler {
	return &Handler{auth: serv}
}

const errInvalidRequestBody = authErr("invalid request body")

type authErr string

func (e authErr) Error() string {
	return string(e)
}

func (h *Handler) SignUp(c *gin.Context) {
	var input model.User
	if err := c.BindJSON(&input); err != nil {
		c.AbortWithError(http.StatusBadRequest, errInvalidRequestBody)
		return
	}

	id, err := h.auth.CreateUser(input)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

func (h *Handler) SignIn(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	token, err := h.auth.GenerateToken(input.Username, input.Password)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}
