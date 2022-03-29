package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const userContext = "userID"

const (
	errEmptyHeader    = authErr("empty authorization header")
	errInvalidHeader  = authErr("invalid authorization header")
	errEmptyToken     = authErr("token is empty")
	errInvalidToken   = authErr("failed to parse token")
	errUserIDNotFound = authErr("user id not found")
	errInvalidUserID  = authErr("user id is of invalid type")
)

func (h *Handler) UserIdentity(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" {
		c.AbortWithError(http.StatusUnauthorized, errEmptyHeader)
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		c.AbortWithError(http.StatusUnauthorized, errInvalidHeader)
		return
	}

	if headerParts[0] != "Bearer" {
		c.AbortWithError(http.StatusUnauthorized, errInvalidHeader)
		return
	}

	if headerParts[1] == "" {
		c.AbortWithError(http.StatusUnauthorized, errEmptyToken)
		return
	}

	userID, err := h.auth.ParseToken(headerParts[1])
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, errInvalidToken)
		return
	}

	c.Set(userContext, userID)
}

func GetUserID(c *gin.Context) (int, error) {
	id, ok := c.Get(userContext)
	if !ok {
		c.AbortWithError(http.StatusInternalServerError, errUserIDNotFound)
		return 0, errUserIDNotFound
	}

	idInt, ok := id.(int)
	if !ok {
		c.AbortWithError(http.StatusInternalServerError, errInvalidUserID)
		return 0, errUserIDNotFound
	}

	return idInt, nil
}
