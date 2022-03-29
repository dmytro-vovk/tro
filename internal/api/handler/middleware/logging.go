package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func ErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	const format = "API error: %s\n"
	return func(c *gin.Context) {
		c.Next()

		entry := logger.WithFields(logrus.Fields{
			//"package": "auth",              // just the sample
			//"function":"(*Handler).SignUp", // just the sample
			//"request":                      // request data (JSON),
			//"user":                         // user id or something else...
			"scheme":   getScheme(c.Request),
			"host":     c.Request.Host,
			"path":     c.Request.URL,
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

func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}

	return "http"
}
