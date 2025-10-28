package main

import (
	"context"
	"os"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Connect to Mongo if MONGO_URI is set
	config.InitMongo(ctx, &cfg)

	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("server start failed")
	}
}
