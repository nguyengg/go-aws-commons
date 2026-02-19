package metrics

import (
	"bytes"
	"log/slog"
	"testing"
	"testing/synctest"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestFactory_NewSlogLogger(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		f := Factory{
			Logger:           &SlogMetricsLogger{},
			ParentSlogLogger: slog.New(slog.NewJSONHandler(&buf, nil)),
		}
		f.String("hello", "world!")
		f.SetCounter("6", 7)
		f.NewSlogLogger().Info("i'm a teapot")

		assert.JSONEq(t, `{"time":"1999-12-31T16:00:00-08:00","level":"INFO","msg":"i'm a teapot","hello":"world!","counters":{"6":7}}`, buf.String())
	})
}

func TestFactory_NewZerologLogger(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)
		f := Factory{
			Logger:              &SlogMetricsLogger{},
			ParentZerologLogger: &logger,
		}
		f.String("hello", "world!")
		f.SetCounter("6", 7)
		f.NewZerologLogger().Info().Msg("i'm a teapot")

		assert.JSONEq(t, `{"level":"info","hello":"world!","counters":{"6":7},"message":"i'm a teapot"}`, buf.String())
	})
}
