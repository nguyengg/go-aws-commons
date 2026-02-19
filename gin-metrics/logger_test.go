package metrics

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gin-gonic/gin"
	commonsmetrics "github.com/nguyengg/go-aws-commons/metrics"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			mbuf bytes.Buffer
			lbuf bytes.Buffer
		)

		logger := slog.New(slog.NewJSONHandler(&lbuf, nil))

		r := gin.New()
		r.Use(Logger(WithRequestId(), WithRecovery(), WithCustomMetrics(func(c *gin.Context) *commonsmetrics.Metrics {
			return commonsmetrics.New(commonsmetrics.LogJSON(&mbuf))
		}), func(cfg *LoggerConfig) {
			cfg.Parent = logger
			cfg.requestId = func() string {
				return "my-request-id"
			}
		}))
		r.GET("/ping", func(c *gin.Context) {
			m := commonsmetrics.Get(c)
			m.AddCounter("userDidSomethingCool", 1)

			time.Sleep(3 * time.Second) // for latency.

			Slog(c).InfoContext(c, "I am the walrus") // expecting this message to contain request Id.

			c.String(http.StatusTeapot, http.StatusText(http.StatusTeapot)+" "+RequestId(c))
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ping", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, 418, w.Code)
		assert.Equal(t, "I'm a teapot my-request-id", w.Body.String())
		assert.JSONEq(t, `{
  "counters": { "fault": 0, "panicked": 0 },
  "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
  "ip": "",
  "latency": "3s",
  "method": "GET",
  "path": "/ping",
  "referrer": "",
  "requestId": "my-request-id",
  "size": 26,
  "startTime": 946684800000,
  "status": 418,
  "userAgent": ""
}`, mbuf.String())
		assert.JSONEq(t, `{
  "time": "1999-12-31T16:00:03-08:00",
  "level": "INFO",
  "msg": "I am the walrus",
  "requestId": "my-request-id"
}`, lbuf.String())
	})
}
