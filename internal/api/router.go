package api

import (
	"github.com/dmytro-vovk/tro/internal/api/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
)

var (
	router = struct {
		once sync.Once
		*gin.Engine
	}{
		Engine: gin.New(),
	}

	log = &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyMsg: "message",
			},
		},
		Hooks:    make(logrus.LevelHooks),
		Level:    logrus.DebugLevel,
		ExitFunc: os.Exit,
	}
)

func (h *Handler) Router() http.Handler {
	router.once.Do(func() {
		router.Use(middleware.Logger(log))

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
	c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "Hello World",
	})
}
