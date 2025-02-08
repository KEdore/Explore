package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/KEdore/explore/internal/config"
	"github.com/KEdore/explore/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a context that cancels on interrupt signals.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	stop, err := server.RunServer(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Block until context is done.
	<-ctx.Done()
	stop()
	log.Println("Server stopped gracefully.")
}
