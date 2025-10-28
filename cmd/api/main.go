package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"archil.lab.tech.com/internal/config"
	"archil.lab.tech.com/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// @title           Go Service API
// @version         1.0
// @description     JSON API with Swagger, WebSockets, Mongo status, and CRUD demo.
// @BasePath        /
// @securityDefinitions.basic BasicAuth
func main() {
	// Structured logging
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(os.Stdout).With().Timestamp().Logger()

	cfg := config.Load()

	// Init Mongo with a bounded timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	config.InitMongo(ctx, &cfg)
	cancel()

	// Ensure Mongo is closed on shutdown
	defer func() {
		cctx, ccancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ccancel()
		config.CloseMongo(cctx, &cfg)
	}()

	// Start HTTP server (your server.Start should return when shutdown finishes)
	srv := server.New(cfg)

	// Handle SIGINT/SIGTERM for graceful shutdown if your server supports it
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownCh
		if stopper, ok := interface{}(srv).(interface{ Shutdown(context.Context) error }); ok {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := stopper.Shutdown(ctx); err != nil {
				log.Error().Err(err).Msg("http shutdown error")
			}
		}
	}()

	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("server start failed")
	}
}
