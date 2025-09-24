# Terminal-Safe Progress Bar

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/tspb.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/tspb)

Provides wrappers around `github.com/schollz/progressbar/v3` that fallbacks to using `log.Logger` if the program does
not have an attached terminal (using `golang.org/x/term` `term.IsTerminal`).
