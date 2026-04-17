package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

var (
	TokensMintedTotal    atomic.Uint64
	VerificationsTotal   atomic.Uint64
	VerificationErrors   atomic.Uint64
	ReplaysDetected      atomic.Uint64
	PolicyDenials        atomic.Uint64
	GatewayRequestsTotal atomic.Uint64
	ReplayCacheSize      atomic.Int64
)

// Handler serves metrics in Prometheus text format.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "# HELP oathmesh_tokens_minted_total Total number of tokens minted\n")
	fmt.Fprintf(w, "# TYPE oathmesh_tokens_minted_total counter\n")
	fmt.Fprintf(w, "oathmesh_tokens_minted_total %d\n", TokensMintedTotal.Load())

	fmt.Fprintf(w, "# HELP oathmesh_verifications_total Total number of token verifications\n")
	fmt.Fprintf(w, "# TYPE oathmesh_verifications_total counter\n")
	fmt.Fprintf(w, "oathmesh_verifications_total %d\n", VerificationsTotal.Load())

	fmt.Fprintf(w, "# HELP oathmesh_verification_errors Total number of token verification validation failures\n")
	fmt.Fprintf(w, "# TYPE oathmesh_verification_errors counter\n")
	fmt.Fprintf(w, "oathmesh_verification_errors %d\n", VerificationErrors.Load())

	fmt.Fprintf(w, "# HELP oathmesh_replays_detected Total number of blocked replay attempts\n")
	fmt.Fprintf(w, "# TYPE oathmesh_replays_detected counter\n")
	fmt.Fprintf(w, "oathmesh_replays_detected %d\n", ReplaysDetected.Load())

	fmt.Fprintf(w, "# HELP oathmesh_policy_denials Total number of policy verification denials\n")
	fmt.Fprintf(w, "# TYPE oathmesh_policy_denials counter\n")
	fmt.Fprintf(w, "oathmesh_policy_denials %d\n", PolicyDenials.Load())

	fmt.Fprintf(w, "# HELP oathmesh_gateway_requests_total Total number of requests proxy through the gateway\n")
	fmt.Fprintf(w, "# TYPE oathmesh_gateway_requests_total counter\n")
	fmt.Fprintf(w, "oathmesh_gateway_requests_total %d\n", GatewayRequestsTotal.Load())

	fmt.Fprintf(w, "# HELP oathmesh_replay_cache_size Current number of tracked jti keys in memory\n")
	fmt.Fprintf(w, "# TYPE oathmesh_replay_cache_size gauge\n")
	fmt.Fprintf(w, "oathmesh_replay_cache_size %d\n", ReplayCacheSize.Load())
}
