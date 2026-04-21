package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TokensMintedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oathmesh_tokens_minted_total",
		Help: "Total number of tokens mints executed securely.",
	})

	VerificationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oathmesh_verifications_total",
		Help: "Total number of tokens formally verified via the 14-step boundary.",
	})

	VerificationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oathmesh_verification_errors",
		Help: "Total number of token evaluations that failed due to validation rules.",
	})

	ReplaysDetected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oathmesh_replays_detected",
		Help: "Total number of identical tokens caught and denied by the replay boundary.",
	})

	PolicyDenials = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oathmesh_policy_denials",
		Help: "Total number of verifications actively rejected by the Pkl security matrix.",
	})

	GatewayRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oathmesh_gateway_requests_total",
		Help: "Total number of requests fully proxied via Gateway instances.",
	})

	ReplayCacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "oathmesh_replay_cache_size",
		Help: "Current total volume of in-memory replay assertions globally tracking.",
	})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "oathmesh_request_duration_seconds",
		Help:    "Latency distributions bound strictly mapping across the runtime.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})

	RevocationSyncErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oathmesh_revocation_sync_errors",
		Help: "Total number of revocation list sync failures",
	})

	GRPCRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "oathmesh_grpc_requests_total",
		Help: "Total number of gRPC requests processed",
	}, []string{"method", "status"})

	GRPCRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "oathmesh_grpc_request_duration_seconds",
		Help:    "gRPC request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})

	ClockSkewRejections = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "oathmesh_clock_skew_rejections_total",
		Help: "Total number of token rejections caused by clock skew between issuer and verifier.",
	}, []string{"reason"})
)

// Handler serves metrics natively adhering to the Prometheus exposition guidelines.
func Handler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}
