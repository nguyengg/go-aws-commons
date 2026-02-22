package metrics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		args time.Duration
		want string
	}{
		{
			name: "5.678s",
			args: time.Duration(5.6789 * float64(time.Second)),
			want: "5.678s",
		},
		{
			name: "535.26ms",
			args: 535_260 * time.Microsecond,
			want: "535.26ms",
		},
		{
			name: "123.456ms",
			args: 123_456_789 * time.Nanosecond,
			want: "123.456ms",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, FormatDuration(tt.args), "FormatSeconds(%v)", tt.args)
		})
	}
}

func TestMetrics_Error(t *testing.T) {
	m := New()
	m.Error(createError())
	m.AnError("myError", createError())

	data, err := m.MarshalJSON()
	assert.NoError(t, err)

	// because data contains local path, can't assert thereon.
	fmt.Printf("%s\n", data)

	if logger := slog.New(slog.NewJSONHandler(os.Stderr, nil)); true {
		logger.LogAttrs(context.Background(), slog.LevelInfo, "", slog.Any("metrics", m))
	}

	if logger := zerolog.New(os.Stderr); true {
		m.e(logger.Info(), false).Send()
	}
}

func createError() error {
	return errors.New("test error")
}
