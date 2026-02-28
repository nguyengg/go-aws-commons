package slogging

import (
	"bytes"
	"log/slog"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
)

func TestJSONStringValue(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "valid JSON",
			args: `{"6": 7}`,
			want: `{"time":"1999-12-31T16:00:00-08:00","level":"INFO","msg":"test","args":{"6":7}}`,
		},
		{
			name: "invalid JSON",
			args: `hello, world!`,
			want: `{"time":"1999-12-31T16:00:00-08:00","level":"INFO","msg":"test","args":"hello, world!"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var buf bytes.Buffer
				logger := slog.New(slog.NewJSONHandler(&buf, nil))
				logger.Info("test", slog.Any("args", JSONStringValue(tt.args)))

				got := buf.String()
				assert.JSONEqf(t, tt.want, got, "JSONStringValue(%v)", tt.args)
			})
		})
	}
}

func TestJSONStringValue_NoJSONHandler(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "valid JSON",
			args: `{"6": 7}`,
			want: "time=1999-12-31T16:00:00.000-08:00 level=INFO msg=test args=\"{\\\"6\\\": 7}\"\n",
		},
		{
			name: "invalid JSON",
			args: `hello, world!`,
			want: "time=1999-12-31T16:00:00.000-08:00 level=INFO msg=test args=\"hello, world!\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var buf bytes.Buffer
				logger := slog.New(slog.NewTextHandler(&buf, nil))
				logger.Info("test", slog.Any("args", JSONStringValue(tt.args)))

				assert.Equalf(t, tt.want, buf.String(), "JSONStringValue(%v)", tt.args)
			})
		})
	}
}

func TestJSONValue(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "valid JSON",
			args: `{"6": 7}`,
			want: `{"time":"1999-12-31T16:00:00-08:00","level":"INFO","msg":"test","args":{"6":7}}`,
		},
		{
			name: "invalid JSON",
			args: `hello, world!`,
			want: `{"time":"1999-12-31T16:00:00-08:00","level":"INFO","msg":"test","args":"aGVsbG8sIHdvcmxkIQ=="}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var buf bytes.Buffer
				logger := slog.New(slog.NewJSONHandler(&buf, nil))
				logger.Info("test", slog.Any("args", JSONValue([]byte(tt.args))))

				got := buf.String()
				assert.JSONEqf(t, tt.want, got, "JSONValue(%v)", tt.args)
			})
		})
	}
}

func TestJSONValue_NoJSONHandler(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "valid JSON",
			args: `{"6": 7}`,
			want: "time=1999-12-31T16:00:00.000-08:00 level=INFO msg=test args=\"eyI2IjogN30=\"\n",
		},
		{
			name: "invalid JSON",
			args: `hello, world!`,
			want: "time=1999-12-31T16:00:00.000-08:00 level=INFO msg=test args=\"hello, world!\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var buf bytes.Buffer
				logger := slog.New(slog.NewTextHandler(&buf, nil))
				logger.Info("test", slog.Any("args", JSONValue([]byte(tt.args))))

				assert.Equalf(t, tt.want, buf.String(), "JSONValue(%v)", tt.args)
			})
		})
	}
}

func TestAnyJSONHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, nil))
		logger.Info("test", slog.Any("obj", struct {
			Test string `json:"test"`
		}{
			Test: "test",
		}))

		assert.JSONEq(t, `{"time":"1999-12-31T16:00:00-08:00","level":"INFO","msg":"test","obj":{"test":"test"}}`, buf.String())
	})
}

func TestAnyTextHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		logger.Info("test", slog.Any("obj", struct {
			Test string `json:"test"`
		}{
			Test: "test",
		}))

		assert.Equal(t, "time=1999-12-31T16:00:00.000-08:00 level=INFO msg=test obj={Test:test}\n", buf.String())
	})
}
