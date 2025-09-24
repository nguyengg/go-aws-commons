package tspb

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

// FromReader creates a new terminal-safe progress bar for reading from the given io.Reader.
//
// If given an os.File, its name and size will be used to provide extra information about read progress. The description
// argument can include `{name}` or `{basename}` which will be replaced with the actual file's name accordingly.
//
// If there is no terminal, an instance of RateLimitedLogger is returned.
func FromReader(r io.Reader, description string) io.ReadCloser {
	var size int64 = -1
	if f, ok := r.(*os.File); ok {
		name := f.Name()
		if fi, err := f.Stat(); err == nil {
			size = fi.Size()
		}

		if description != "" {
			description = strings.ReplaceAll(description, "{name}", name)
			description = strings.ReplaceAll(description, "{basename}", filepath.Base(name))
		} else {
			description = fmt.Sprintf(`reading from "%s"`, filepath.Base(name))
		}
	} else if description == "" {
		description = "reading"
	}

	if term.IsTerminal(int(os.Stdout.Fd())) {
		pb := progressbar.NewReader(r, progressbar.NewOptions64(size, defaultByteOptions(description)...))
		return &pb
	}

	return newRateLimitedLogger(description, size)
}

// FromReaderWithOptions is a variant of FromReader that gives caller more customisation options over the progress bar.
func FromReaderWithOptions(r io.Reader, options ...progressbar.Option) io.ReadCloser {
	var (
		description string
		size        int64 = -1
	)

	if f, ok := r.(*os.File); ok {
		name := f.Name()
		if fi, err := f.Stat(); err == nil {
			size = fi.Size()
		}

		description = fmt.Sprintf(`reading from "%s"`, filepath.Base(name))
	} else {
		description = "reading"
	}

	if term.IsTerminal(int(os.Stdout.Fd())) {
		options = append([]progressbar.Option{progressbar.OptionSetDescription(description)}, options...)
		pb := progressbar.NewReader(r, progressbar.NewOptions64(size, options...))
		return &pb
	}

	return newRateLimitedLogger(description, size)
}
