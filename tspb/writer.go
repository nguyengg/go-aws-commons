package tspb

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// FromWriter creates a new terminal-safe progress bar for writing to the given io.Writer.
//
// If given an os.File, its name and size will be used to provide extra information about read progress. The description
// argument can include `{name}` or `{basename}` which will be replaced with the actual file's name accordingly.
//
// If there is no terminal, an instance of RateLimitedLogger is returned.
func FromWriter(w io.Writer, description string) io.WriteCloser {
	var size int64 = -1
	if f, ok := w.(*os.File); ok {
		name := f.Name()
		if fi, err := f.Stat(); err == nil {
			size = fi.Size()
		}

		if description != "" {
			description = strings.ReplaceAll(description, "{name}", name)
			description = strings.ReplaceAll(description, "{basename}", filepath.Base(name))
		} else {
			description = fmt.Sprintf(`writing to "%s"`, filepath.Base(name))
		}
	} else if description == "" {
		description = "writing"
	}

	return DefaultBytes(size, description)
}

// FromWriterWithOptions is a variant of FromWriter that gives caller more customisation options over the progress bar.
func FromWriterWithOptions(w io.Writer, options ...progressbar.Option) io.WriteCloser {
	var (
		description string
		size        int64 = -1
	)

	if f, ok := w.(*os.File); ok {
		name := f.Name()
		if fi, err := f.Stat(); err == nil {
			size = fi.Size()
		}

		description = fmt.Sprintf(`writing to "%s"`, filepath.Base(name))
	} else {
		description = "writing"
	}

	return DefaultBytes(size, description, options...)
}
