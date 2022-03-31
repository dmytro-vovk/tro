package api

import (
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"

	"github.com/dmytro-vovk/tro/internal/api/handler/middleware"
	"github.com/gin-gonic/gin"
)

var router = struct {
	once sync.Once
	*gin.Engine
}{
	Engine: gin.New(),
}

func (h *Handler) Router() http.Handler {
	router.once.Do(func() {
		logrus.SetLevel(logrus.DebugLevel)
		//logrus.SetFormatter(&logrus.JSONFormatter{
		//	PrettyPrint: true,
		//})
		router.Use(middleware.Logger)

		auth := router.Group("/auth")
		{
			auth.POST("/sign-up", h.auth.SignUp)
			auth.POST("/sign-in", h.auth.SignIn)
		}

		//api := router.Group("/api", h.auth.UserIdentity)
		api := router.Group("/api")
		{
			api.GET("/hello-world", h.helloWorld)
			api.POST("/hello-world", h.helloWorld)
		}
	})

	return router.Engine
}

func (h *Handler) helloWorld(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		var input struct {
			Hello string `json:"hello" binding:"required"`
			World string `json:"world"`
		}
		if err := c.BindJSON(&input); err != nil {
			c.Error(errors.New("non-fatal error")).SetMeta(map[string]interface{}{
				"fuel":      "diesel",
				"transport": 2,
			})

			c.AbortWithError(http.StatusBadRequest, errors.New("fatal error")).SetMeta(map[string]interface{}{
				"meta": "context",
				"test": 999,
			})
			return
		}
	}

	c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "Hello World",
	})
}
