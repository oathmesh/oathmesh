package issuer

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/MustafaMahmoudAtta111/oathmesh/internal/sign"
)

type Server struct {
	httpServer *http.Server
	keySet     interface {
		GetIssuer() string
		JWKS() (*sign.JWKS, error)
		SignToken(sign.MintRequest) (string, error)
	}
	logger         *slog.Logger
	port           string
	rateLimiter    *RateLimiter
	gatewayHandler http.Handler
}

func NewServer(keySet interface {
	GetIssuer() string
	JWKS() (*sign.JWKS, error)
	SignToken(sign.MintRequest) (string, error)
}) *Server {
	port := os.Getenv("OATHMESH_PORT")
	if port == "" {
		port = "4000"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	rateLimiter := NewRateLimiter(100, 20)

	return &Server{
		keySet:      keySet,
		logger:      logger,
		port:        port,
		rateLimiter: rateLimiter,
	}
}

// SetGateway enables reverse proxy mode by setting a handler for all
// paths not explicitly registered by the issuer.
func (s *Server) SetGateway(h http.Handler) {
	s.gatewayHandler = h
}

func (s *Server) Run() error {
	r := s.router()

	s.httpServer = &http.Server{
		Addr:         ":" + s.port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("starting issuer server", "port", s.port)
	if s.gatewayHandler != nil {
		s.logger.Info("gateway mode enabled")
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		s.logger.Info("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}
}

func (s *Server) router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middlewareLogger(s.logger))
	r.Use(middleware.Recoverer)

	r.Get("/healthz", healthzHandler)

	r.Group(func(r chi.Router) {
		r.Get("/.well-known/jwks.json", s.jwksHandler)
		r.Get("/.well-known/oathmesh-issuer", s.discoveryHandler)
	})

	r.Group(func(r chi.Router) {
		r.Post("/v1/token", s.mintHandler)
		r.Post("/v1/exchange/github", s.exchangeGitHubHandler)
	})

	if s.gatewayHandler != nil {
		r.NotFound(s.gatewayHandler.ServeHTTP)
	}

	return r
}

func middlewareLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start),
			)
		})
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
