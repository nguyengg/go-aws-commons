package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/nguyengg/go-aws-commons/tspb"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stop()

	bar := tspb.New(-1, "test")
	defer bar.Close()

	buf := make([]byte, 32)

	for ticker := time.NewTicker(1 * time.Second); ; {
		select {
		case <-ctx.Done():
			_ = bar.Close()
			log.Printf("interrupted")
			return
		case <-ticker.C:
			_, _ = bar.Write(buf)
		}
	}
}
