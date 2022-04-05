package api

import (
	"github.com/dmytro-vovk/tro/internal/api/handler/middleware"
	"github.com/gin-gonic/gin"
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

func (h *Handler) helloWorld(c *gin.Context) {
	h.log.Debug("Debug message")
	h.log.Info("Info message")
	h.log.Warning("Warning message")
	h.log.Error("Error message")
	c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "Hello World",
	})
}
