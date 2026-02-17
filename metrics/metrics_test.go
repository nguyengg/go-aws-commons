package metrics

import (
	"testing"
	"time"

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
