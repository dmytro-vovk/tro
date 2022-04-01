package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func Logger(log *logrus.Logger) gin.HandlerFunc {
	if f, ok := log.Formatter.(*logrus.TextFormatter); ok {
		f.ForceQuote = false
		f.DisableQuote = true
	}

	return func(c *gin.Context) {
		req, err := getRequest(c, log.WithFields(logrus.Fields{
			"client_ip": c.ClientIP(),
			"user_id":   "1",
		}))
		if err != nil {
			req.withFields(logrus.Fields{
				"package":  "middleware",
				"function": "getRequest",
			}).Errorln("Can't get log request:", err)
			return
		}

		c.Next()

		req.handleResponse(c)
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
	rwt       *responseWriter
	log       *logrus.Entry
}

func getRequest(c *gin.Context, e *logrus.Entry) (*request, error) {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	r := &request{
		UserAgent: c.Request.UserAgent(),
		Scheme:    scheme,
		Host:      c.Request.Host,
		Path:      c.Request.URL.EscapedPath(),
		Method:    c.Request.Method,
		rwt:       newResponseWriter(c),
		log:       e,
	}
	r.withField("request", r)

	payload, err := c.GetRawData()
	if err != nil {
		return r, err
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(payload))
	r.Body = getBody(payload)
	return r, nil
}

// String used for pretty printing in logrus.Logger when set logrus.TextFormatter
func (r *request) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		r.log.Panic(err)
	}
	return string(b)
}

func (r *request) withField(key string, value interface{}) *logrus.Entry {
	return r.withFields(logrus.Fields{key: value})
}

func (r *request) withFields(fields logrus.Fields) *logrus.Entry {
	r.log = r.log.WithFields(fields)
	return r.log
}

type apiErr struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// handleErrors must be called before handleResponse
func (r *request) handleErrors(c *gin.Context) error {
	code := c.Writer.Status()
	if code < http.StatusBadRequest {
		return nil
	}

	if l := len(c.Errors); l >= 2 {
		for _, ginErr := range c.Errors[:l-1] {
			r.log.Infoln("API preceding error:", ginErr)
		}
	}

	if err := c.Errors.Last(); err != nil {
		c.JSON(code, gin.H{
			"error": apiErr{
				Code:    code,
				Status:  http.StatusText(code),
				Message: err.Error(),
				Details: err.Meta,
			},
		})
		return err
	}

	return nil
}

// handleResponse must be called after c.Next()
func (r *request) handleResponse(c *gin.Context) {
	err := r.handleErrors(c)

	resp := &response{
		Code: c.Writer.Status(),
		Body: getBody(r.rwt.body.Bytes()),
		log:  r.log,
	}

	r.withField("response", resp)
	switch {
	case resp.isInfo(), resp.isSuccess(), resp.isRedirect():
		r.log.Debug("API success")
	case resp.isClientError():
		r.log.Warningln("API error:", err)
	case resp.isServerError():
		r.log.Errorln("API error:", err)
	}
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

type response struct {
	Code int         `json:"code"`
	Body interface{} `json:"body,omitempty"`
	log  *logrus.Entry
}

// String used for pretty printing in logrus.Logger when set logrus.TextFormatter
func (r *response) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		r.log.Panic(err)
	}
	return string(b)
}

func (r *response) isInfo() bool {
	return r.Code >= http.StatusContinue && r.Code <= http.StatusEarlyHints
}

func (r *response) isSuccess() bool {
	return r.Code >= http.StatusOK && r.Code <= http.StatusAlreadyReported ||
		r.Code == http.StatusIMUsed
}

func (r *response) isRedirect() bool {
	return r.Code >= http.StatusMultipleChoices && r.Code <= http.StatusPermanentRedirect
}

func (r *response) isClientError() bool {
	return r.Code >= http.StatusBadRequest && r.Code <= http.StatusTeapot ||
		r.Code >= http.StatusMisdirectedRequest && r.Code <= http.StatusUpgradeRequired ||
		r.Code == http.StatusPreconditionRequired ||
		r.Code == http.StatusTooManyRequests ||
		r.Code == http.StatusRequestHeaderFieldsTooLarge ||
		r.Code == http.StatusUnavailableForLegalReasons
}

func (r *response) isServerError() bool {
	return r.Code >= http.StatusInternalServerError && r.Code <= http.StatusNetworkAuthenticationRequired
}

// getBody used for pretty printing in logrus.Logger
func getBody(payload []byte) interface{} {
	if json.Valid(payload) {
		return json.RawMessage(payload)
	}

	return string(payload)
}
