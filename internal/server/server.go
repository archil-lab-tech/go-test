package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"archil.lab.tech.com/internal/config"
	"archil.lab.tech.com/internal/handlers"
	"archil.lab.tech.com/internal/mid"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type Server struct {
	cfg config.Config
	srv *http.Server
}

func New(cfg config.Config) *Server {
	r := chi.NewRouter()

	// Core middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mid.RecoverJSON)
	r.Use(mid.Logger)

	// Public routes
	r.Get("/", handlers.RootJSON)
	r.Get("/healthz", handlers.Healthz)
	r.Get("/readyz", handlers.Readyz)

	// WS demo UI and endpoint
	r.Get("/ui", handlers.UI) // HTML page
	r.Get("/ws", handlers.WS) // WebSocket JSON stream

	// Status (Mongo connectivity JSON)
	r.Get("/status", handlers.Status)

	r.Get("/debug/env", handlers.EnvProbe)

	// CRUD demo
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/items", handlers.ItemsList)
		api.Post("/items", handlers.ItemCreate)
		api.Get("/items/{id}", handlers.ItemGet)
		api.Put("/items/{id}", handlers.ItemPut)
		api.Patch("/items/{id}", handlers.ItemPatch)
		api.Delete("/items/{id}", handlers.ItemDelete)
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	s := &http.Server{
		// Cloud Run injects PORT; cfg.Port already contains the numeric string (e.g., "8080")
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return &Server{cfg: cfg, srv: s}
}

func (s *Server) Start() error {
	// Share cfg with handlers
	handlers.SetRuntime(&s.cfg)

	errCh := make(chan error, 1)

	// Start HTTP listener
	go func() {
		log.Info().Str("addr", s.srv.Addr).Msg("http: listening")
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		close(errCh)
	}()

	// Handle SIGINT/SIGTERM for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Info().Str("sig", sig.String()).Msg("http: shutdown signal received")
	case err := <-errCh:
		// Listen error bubbled up
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Graceful HTTP shutdown
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("http: shutdown error")
	}

	// Close Mongo (safe if nil)
	config.CloseMongo(ctx, &s.cfg)

	log.Info().Msg("http: stopped")
	return nil
}
