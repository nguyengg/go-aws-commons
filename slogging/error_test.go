package slogging

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"testing/synctest"
)

func TestErrorValue_Text(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		logger.Info("test", AnError("error", gimmeError()))

		// unfortunately this will contain the path of the file so can't assert on this.
		fmt.Print(buf.String())
	})
}

func TestErrorValue_JSON(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, nil))
		logger.Info("test", AnError("error", gimmeError()))

		// unfortunately this will contain the path of the file so can't assert on this.
		fmt.Print(buf.String())
	})
}

func gimmeError() error {
	return errors.New("test error")
}
