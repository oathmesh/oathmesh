package issuer

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"

	"github.com/oathmesh/oathmesh/internal/config"
	"github.com/oathmesh/oathmesh/internal/metrics"
	"github.com/oathmesh/oathmesh/internal/sign"
	"github.com/oathmesh/oathmesh/internal/verify"
)

// maxRequestBodySize is the maximum allowed request body for POST endpoints.
// 64 KiB is more than sufficient for any OathMesh mint or exchange request.
const maxRequestBodySize = 64 * 1024

type Server struct {
	httpServer *http.Server
	keySet         sign.Signer
	logger         *slog.Logger
	port           string
	cfg            *config.Config
	rateLimiter    *RateLimiter
	gatewayHandler http.Handler
	revocations    *verify.RedisRevocationList
}

func NewServer(keySet sign.Signer) *Server {
	cfg := config.LoadFromEnv()

	port := os.Getenv("OATHMESH_PORT")
	if port == "" {
		port = "4000"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	rateLimiter := NewRateLimiter(cfg.RateLimitRPM, cfg.RateLimitBurst)

	var revList *verify.RedisRevocationList
	if cfg.RedisURL != "" {
		rl, err := verify.NewRedisRevocationList(cfg.RedisURL)
		if err == nil {
			revList = rl
		}
	}

	return &Server{
		keySet:      keySet,
		logger:      logger,
		port:        port,
		cfg:         cfg,
		rateLimiter: rateLimiter,
		revocations: revList,
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
	r.Use(otelchi.Middleware("oathmesh-issuer"))
	r.Use(metrics.LatencyMiddleware)

	// ── Public endpoints (no auth) ──────────────────────────────────────
	r.Get("/healthz", healthzHandler)
	r.Get("/metrics", metrics.Handler)

	r.Group(func(r chi.Router) {
		r.Get("/.well-known/jwks.json", s.jwksHandler)
		r.Get("/.well-known/oathmesh-issuer", s.discoveryHandler)
		r.Get("/v1/revoked-subjects", s.revokedSubjectsHandler)
	})

	// ── Authenticated mint & admin endpoints ───────────────────────────
	// Protected by MintAuth middleware (pre-shared key).
	r.Group(func(r chi.Router) {
		r.Use(MintAuth)
		r.Post("/v1/token", s.mintHandler)
		r.Post("/v1/exchange/github", s.exchangeGitHubHandler)
		r.Post("/v1/exchange/gitlab", s.exchangeGitLabHandler)
		r.Post("/v1/admin/revoke", s.revokeHandler)
		r.Delete("/v1/admin/revoke", s.unrevokeHandler)
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
	_, _ = w.Write([]byte("ok"))
}

// ── Revocation Handlers ──────────────────────────────────────────────────────

type RevokeRequest struct {
	Subject string `json:"sub"`
}

func (s *Server) revokeHandler(w http.ResponseWriter, r *http.Request) {
	if s.revocations == nil {
		s.writeError(w, "revocation_disabled", "Redis is not configured. Revocation is unavailable.", "Set REDIS_URL")
		return
	}

	var req RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "invalid_request", "Failed to parse request body", "")
		return
	}
	if req.Subject == "" {
		s.writeError(w, "missing_subject", "Subject is required", "")
		return
	}

	if err := s.revocations.Revoke(r.Context(), req.Subject); err != nil {
		s.logger.Error("failed to revoke subject", "error", err)
		s.writeError(w, "internal_error", "Failed to revoke subject", "")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "revoked"}`))
}

func (s *Server) unrevokeHandler(w http.ResponseWriter, r *http.Request) {
	if s.revocations == nil {
		s.writeError(w, "revocation_disabled", "Redis is not configured. Revocation is unavailable.", "Set REDIS_URL")
		return
	}

	var req RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "invalid_request", "Failed to parse request body", "")
		return
	}
	if req.Subject == "" {
		s.writeError(w, "missing_subject", "Subject is required", "")
		return
	}

	if err := s.revocations.Unrevoke(r.Context(), req.Subject); err != nil {
		s.logger.Error("failed to unrevoke subject", "error", err)
		s.writeError(w, "internal_error", "Failed to unrevoke subject", "")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "unrevoked"}`))
}

func (s *Server) revokedSubjectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.revocations == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"revocations": []}`))
		return
	}

	list, err := s.revocations.List(r.Context())
	if err != nil {
		s.logger.Error("failed to list revocations", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"revocations": list,
	})
}
