package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/MustafaMahmoudAtta111/oathmesh/sdk/go/middleware"
	"github.com/MustafaMahmoudAtta111/oathmesh/internal/verify"
	"github.com/MustafaMahmoudAtta111/oathmesh/internal/policy"
	"github.com/MustafaMahmoudAtta111/oathmesh/internal/audit"
)

func main() {
	audience := os.Getenv("OATHMESH_AUDIENCE")
	trustedIssuersStr := os.Getenv("OATHMESH_TRUSTED_ISSUERS")
	policyPath := os.Getenv("OATHMESH_POLICY_PATH")

	if audience == "" || trustedIssuersStr == "" {
		log.Fatal("OATHMESH_AUDIENCE and OATHMESH_TRUSTED_ISSUERS must be set")
	}

	issuers := strings.Split(trustedIssuersStr, ",")
	for i, v := range issuers {
		issuers[i] = strings.TrimSpace(v)
	}

	var evaluator verify.PolicyEvaluator
	if policyPath != "" {
		pe, err := policy.NewWatchedPolicyEngine(policyPath, slog.Default())
		if err != nil {
			log.Fatalf("failed to init policy engine: %v", err)
		}
		evaluator = pe
	}

	cfg := &verify.VerifierConfig{
		Audience:        audience,
		TrustedIssuers:  issuers,
		JWKSProvider:    verify.NewJWKSCache(verify.DefaultJWKSCacheTTL),
		ReplayCache:     verify.NewMemoryReplayCache(),
		PolicyEvaluator: evaluator,
		AuditSink:       audit.NewStdoutAuditSink(),
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.OathMeshMiddleware(cfg))

		r.Get("/inventory", func(w http.ResponseWriter, r *http.Request) {
			caller := middleware.CallerFrom(r.Context())
			if caller == nil {
				http.Error(w, "internal error: caller context missing", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data":   []string{"item1", "item2"},
				"caller": caller,
			})
		})
	})

	log.Println("Starting examples/chi-api on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
