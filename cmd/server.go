package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/JpUnique/TinyGo/pkg/api"
	"github.com/JpUnique/TinyGo/pkg/service"
	"github.com/JpUnique/TinyGo/pkg/store"
)

type Server struct {
	httpServer *http.Server
}

// New creates and configures the HTTP server
func New() (*Server, error) {
	// Load env
	pgURL := os.Getenv("POSTGRES_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Init Postgres
	pg, err := store.NewPostgres(pgURL)
	if err != nil {
		return nil, err
	}

	// Init Redis
	rd, err := store.NewRedis(redisAddr)
	if err != nil {
		return nil, err
	}

	// Init service
	shortenerSvc := service.NewShortener(pg, rd)

	// Router
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API Handlers
	urlHandler := api.NewURLHandler(shortenerSvc)
	urlHandler.RegisterRoutes(r)

	// HTTP Server
	httpSrv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{httpServer: httpSrv}, nil
}

// Start runs the HTTP server and sets up graceful shutdown
func (s *Server) Start() error {
	go func() {
		log.Printf("Server running on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}
