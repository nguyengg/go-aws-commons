# Terminal-Safe Progress Bar

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/tspb.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/tspb)

Provides wrapper around [github.com/schollz/progressbar/v3](https://pkg.go.dev/github.com/schollz/progressbar/v3) that
fallbacks to using `log.Default` if the program does not have an attached terminal to `os.Stderr` ([term.IsTerminal](https://pkg.go.dev/golang.org/x/term#IsTerminal)).
