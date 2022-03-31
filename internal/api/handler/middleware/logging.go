package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

type logWriter struct {
	gin.ResponseWriter
	*logrus.Entry
	req      map[string]interface{}
	resp     map[string]interface{}
	respBody *bytes.Buffer
}

// newLogWriter must be called before c.Next()
func newLogWriter(c *gin.Context, e *logrus.Entry) (*logWriter, error) {
	lw := &logWriter{
		ResponseWriter: c.Writer,
		Entry:          e,
		req: map[string]interface{}{
			"scheme": getScheme(c),
			"host":   c.Request.URL.Hostname(),
			"path":   c.Request.URL.EscapedPath(),
			"method": c.Request.Method,
		},
		resp:     map[string]interface{}{},
		respBody: &bytes.Buffer{},
	}
	lw.withField("request", lw.req)

	body, err := getPayload(c)
	if err != nil {
		return nil, err
	}
	lw.req["body"] = jsonify(body)

	c.Writer = lw
	return lw, nil
}

func (w logWriter) Write(b []byte) (int, error) {
	w.respBody.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w logWriter) WriteString(s string) (int, error) {
	w.respBody.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *logWriter) withFields(fields logrus.Fields) *logrus.Entry {
	w.Entry = w.Entry.WithFields(fields)
	return w.Entry
}

func (w *logWriter) withField(key string, value interface{}) *logrus.Entry {
	w.Entry = w.Entry.WithField(key, value)
	return w.Entry
}

func Logger(logger *logrus.Logger) gin.HandlerFunc {
	//logger.SetFormatter(&logrus.JSONFormatter{
	//	PrettyPrint: true,
	//})
	logger.SetLevel(logrus.DebugLevel)

	return func(c *gin.Context) {
		lw, err := newLogWriter(c, logger.WithFields(logrus.Fields{
			"client_ip": c.ClientIP(),
			"user_id":   "1",
		}))
		if err != nil {
			lw.WithFields(logrus.Fields{
				"package":  "middleware",
				"function": "newLogWriter",
			}).Errorf("Can't get log writer: %s", err)
			return
		}

		start := time.Now()
		c.Next()

		code := c.Writer.Status()
		lw.resp["code"] = code

		if code >= 400 {
			const format = "API error: %s\n"
			if l := len(c.Errors); l >= 2 {
				for _, ginErr := range c.Errors[:l-2] {
					lw.Warningf(format, ginErr)
				}
			}

			if err := c.Errors.Last(); err != nil {
				c.JSON(code, err)
				lw.withField("latency", time.Since(start).String())
				lw.resp["body"] = jsonify(lw.respBody.Bytes())
				lw.withField("response", lw.resp).Errorf(format, err)
			}

			return
		}

		lw.withField("latency", time.Since(start).String())
		lw.resp["body"] = jsonify(lw.respBody.Bytes())
		lw.Debug("Success")
	}
}

func getScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}

	return "http"
}

// getPayload must be called before c.Next()
func getPayload(c *gin.Context) ([]byte, error) {
	payload, err := c.GetRawData()
	if err != nil {
		return nil, err
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(payload))
	return payload, nil
}

func jsonify(payload []byte) interface{} {
	var obj interface{}
	if err := json.Unmarshal(payload, &obj); err != nil {
		return string(payload)
	}

	return obj
}
