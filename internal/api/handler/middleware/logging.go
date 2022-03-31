package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	const format = "API error: %s\n"
	return func(c *gin.Context) {
		c.Next()

		entry := logger.WithFields(logrus.Fields{
			"path":     c.Request.URL.Path,
			"method":   c.Request.Method,
			"response": c.Writer.Status(),
		})

		if l := len(c.Errors); l >= 2 {
			for _, ginErr := range c.Errors[:l-2] {
				entry.Warningf(format, ginErr)
			}
		}

		if err := c.Errors.Last(); err != nil {
			entry.Errorf(format, err)
			c.JSON(-1, err)
		}
	}
}
