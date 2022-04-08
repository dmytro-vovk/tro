package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type errString string

func (e errString) Error() string {
	return string(e)
}

const errInvalidRequestBody errString = "invalid request body"

func Logger(log *logrus.Logger) gin.HandlerFunc {
	if f, ok := log.Formatter.(*logrus.TextFormatter); ok {
		f.ForceQuote = false
		f.DisableQuote = true
	}

	return func(c *gin.Context) {
		req, err := newRequest(c)
		entry := log.WithFields(logrus.Fields{
			"client_ip": c.ClientIP(),
			"user_id":   "1",
			"request":   req,
		})

		defer newResponseWriter(c).handle(c, entry)

		if err != nil {
			entry.Warningln("Can't log request:", err)
			c.AbortWithError(http.StatusBadRequest, errInvalidRequestBody)
			return
		}

		c.Next()
	}
}

type request struct {
	ID        string      `json:"id"` // todo: implement
	UserAgent string      `json:"user_agent"`
	Scheme    string      `json:"scheme"`
	Host      string      `json:"host"`
	Path      string      `json:"path"`
	Method    string      `json:"method"`
	Body      interface{} `json:"body,omitempty"`
}

func newRequest(c *gin.Context) (*request, error) {
	r := &request{
		UserAgent: c.Request.UserAgent(),
		Scheme:    getRequestScheme(c),
		Host:      c.Request.Host,
		Path:      c.Request.URL.EscapedPath(),
		Method:    c.Request.Method,
	}

	body, err := getRequestBody(c)
	if err != nil {
		return r, err
	}

	r.Body = prettify(body)
	return r, nil
}

func (r *request) String() string { return jsonify(r) }

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

func (w *responseWriter) handle(c *gin.Context, entry *logrus.Entry) {
	code := c.Writer.Status()

	if err := c.Errors.Last(); code >= http.StatusBadRequest && err != nil {
		for i, ginErr := range c.Errors {
			// todo: mb do something with ginErr.Meta
			entry.Infof("%d of %d error in chain: %s", i+1, len(c.Errors), ginErr)
		}

		c.JSON(code, gin.H{
			"error": apiErr{
				Code:    code,
				Status:  http.StatusText(code),
				Message: err.Error(),
				Details: err.Meta,
			},
		})
	}

	entry.WithField("response", &response{
		Code: code,
		Body: prettify(w.body.Bytes()),
	}).Info("Request processed")
}

type apiErr struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e apiErr) Error() string {
	return e.Message
}

type response struct {
	Code int         `json:"code"`
	Body interface{} `json:"body,omitempty"`
}

func (r *response) String() string { return jsonify(r) }

func getRequestScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}

	return "http"
}

func getRequestBody(c *gin.Context) ([]byte, error) {
	payload, err := c.GetRawData()
	if err != nil {
		return nil, errors.Wrap(err, "can't read request body")
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(payload))
	return payload, nil
}

func prettify(payload []byte) interface{} {
	if json.Valid(payload) {
		return json.RawMessage(payload)
	}

	return string(payload)
}

func jsonify(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		logrus.Panic(err)
	}
	return string(b)
}
