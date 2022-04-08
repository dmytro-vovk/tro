package api

import (
	"errors"
	"github.com/dmytro-vovk/tro/internal/api/handler/middleware"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"sync"
)

var router = struct {
	*gin.Engine
	once sync.Once
}{
	Engine: gin.New(),
}

func (h *Handler) Router() http.Handler {
	router.once.Do(func() {
		router.Use(middleware.Logger(h.log))

		auth := router.Group("/auth")
		{
			auth.POST("/sign-up", h.auth.SignUp)
			auth.POST("/sign-in", h.auth.SignIn)
		}

		//api := router.Group("/api", h.auth.UserIdentity)
		api := router.Group("/api")
		{
			api.GET("/hello-world", h.helloWorld)
		}
	})

	return router.Engine
}

var counter = 0

func (h *Handler) helloWorld(c *gin.Context) {
	buf := make([]byte, 0, 64)
	for i := 0; i < cap(buf); i++ {
		buf = append(buf, byte(65+rand.Intn(26)))
	}

	h.log.Debug("Debug message")
	h.log.Info("Info message")
	if counter > 1 {
		c.Error(errors.New("first error")).SetMeta(gin.H{"context": "first"})
		c.AbortWithError(http.StatusInternalServerError, errors.New("second error")).SetMeta(gin.H{"context": "second"})
		counter++
		return
	}

	counter++
	c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: string(buf),
	})
}
