package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type request struct {
	ID        string `json:"id"` // todo: implement
	UserAgent string `json:"user_agent"`
	Scheme    string `json:"scheme"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	Body      string `json:"body,omitempty"`
}

func newRequest(c *gin.Context) (*request, error) {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	req := &request{
		UserAgent: c.Request.UserAgent(),
		Scheme:    scheme,
		Host:      c.Request.Host,
		Path:      c.Request.URL.EscapedPath(),
		Method:    c.Request.Method,
	}

	payload, err := c.GetRawData()
	if err != nil {
		return req, err
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(payload))
	req.Body = string(payload)
	return req, nil
}

// todo: mb should implement string method
type response struct {
	Code int    `json:"code"`
	Body string `json:"body,omitempty"`
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func newResponseWriter(c *gin.Context) *responseWriter {
	w := &responseWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
	}
	c.Writer = w
	return w
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func Logger(c *gin.Context) {
	req, err := newRequest(c)
	entry := logrus.WithFields(logrus.Fields{
		"client_ip": c.ClientIP(),
		"user_id":   "1",
		"request":   req,
	})
	if err != nil {
		entry.WithFields(logrus.Fields{
			"package":  "middleware",
			"function": "newRequest",
		}).Error("Can't get request body:", err)
		return
	}

	// responseWriter must be declared before c.Next() for copying data from the c.Writer
	// but the response code we will get only when the chain stop
	w := newResponseWriter(c)
	c.Next()
	code := c.Writer.Status()

	if code >= http.StatusBadRequest {
		if l := len(c.Errors); l >= 2 {
			for _, ginErr := range c.Errors[:l-1] {
				entry.Info("API preceding error: ", ginErr)
			}
		}

		if err := c.Errors.Last(); err != nil {
			c.JSON(code, err)
		}
	}

	entry = entry.WithField("response", &response{
		Code: code,
		Body: w.body.String(),
	})
	switch {
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		entry.Warning("API error: ", err)
	case code >= http.StatusInternalServerError:
		entry.Error("API error: ", err)
	default:
		entry.Debug("API success")
	}
}
