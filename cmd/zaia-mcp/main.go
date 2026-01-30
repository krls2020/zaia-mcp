package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zeropsio/zaia-mcp/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	srv := server.New()
	if err := srv.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
