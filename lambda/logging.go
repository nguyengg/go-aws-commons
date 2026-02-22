package lambda

import (
	"log"
	"log/slog"
	"os"
	"strconv"
)

const (
	// DebugLogFlags is the flag passed to log.SetFlags if DEBUG environment variable is true-ish.
	DebugLogFlags = log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Llongfile | log.Lmsgprefix

	// DefaultLogFlags is the flag passed to log.SetFlags if DEBUG environment is not true-ish.
	DefaultLogFlags = log.LstdFlags | log.LUTC | log.Lmsgprefix
)

// SetUpLogDefault sets up flags for the default logger depending on the DEBUG environment variable.
//
// If the DEBUG environment variable is true-ish, DebugLogFlags is passed to log.SetFlags. Otherwise, DefaultLogFlags is
// passed to log.SetFlags.
func SetUpLogDefault() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.SetFlags(DebugLogFlags)
	} else {
		log.SetFlags(DefaultLogFlags)
	}
}

// SetUpSlogDefault sets the default slog.Default to print JSON contents to os.Stderr.
//
// Additionally, if DEBUG environment variable is true-ish, slog.SetLogLoggerLevel is set to slog.LevelDebug prior to
// the slog.SetDefault call, and the slog.JSONHandler will output messages at slog.LevelDebug level. This also has the
// effect of making every subsequent log.Printf to also print at slog.LevelDebug. See slog.SetLogLoggerLevel for a more
// in-depth explanation.
func SetUpSlogDefault() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	var level slog.Level
	if debug {
		level = slog.LevelDebug
		slog.SetLogLoggerLevel(level)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}
