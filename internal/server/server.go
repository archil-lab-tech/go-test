package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"archil.lab.tech.com/internal/common"
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

	// Make runtime config available to handlers ASAP
	common.SetRuntime(&cfg)
	// (Remove this duplicate to avoid confusion)
	// handlers.SetRuntime(&cfg)

	// Core middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mid.RecoverJSON)
	r.Use(mid.Logger)

	// Public JSON routes
	r.Get("/", handlers.RootJSON)
	r.Get("/healthz", handlers.Healthz)
	r.Get("/readyz", handlers.Readyz)
	r.Get("/status", handlers.Status)

	// Debug helpers
	r.Get("/debug/env", handlers.EnvProbe)
	r.Get("/egress-ip", handlers.EgressIP)
	r.Get("/debug/mongo", handlers.MongoDiag)

	// WS demo + Swagger (non-JSON)
	r.Get("/ui", handlers.UI) // HTML
	r.Get("/ws", handlers.WS) // WebSocket
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// CRUD demo (JSON)
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/items", handlers.ItemsList)
		api.Post("/items", handlers.ItemCreate)
		api.Get("/items/{id}", handlers.ItemGet)
		api.Put("/items/{id}", handlers.ItemPut)
		api.Patch("/items/{id}", handlers.ItemPatch)
		api.Delete("/items/{id}", handlers.ItemDelete)
	})

	// Force JSON for 404/405 coming through your app
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		common.WriteJSON(w, http.StatusNotFound, map[string]any{
			"ok": false, "error": "not_found", "path": r.URL.Path,
		})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		common.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"ok": false, "error": "method_not_allowed", "method": r.Method, "path": r.URL.Path,
		})
	})

	s := &http.Server{
		Addr:              ":" + cfg.Port, // Cloud Run injects PORT
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return &Server{cfg: cfg, srv: s}
}

func (s *Server) Start() error {
	errCh := make(chan error, 1)

	go func() {
		log.Info().Str("addr", s.srv.Addr).Msg("http: listening")
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		close(errCh)
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Info().Str("sig", sig.String()).Msg("http: shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("http: shutdown error")
	}

	log.Info().Msg("http: stopped")
	return nil
}
