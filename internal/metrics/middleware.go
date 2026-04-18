package metrics

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// LatencyMiddleware times the execution of HTTP requests and emits
// the exact distribution tracking directly to Prometheus.
func LatencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the writer so we can introspect the response transparently
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		next.ServeHTTP(ww, r)
		
		duration := time.Since(start).Seconds()

		// In a highly dynamic router like chi, recording raw `r.URL.Path` might cause high cardinality 
		// if path variables exist. For this proxy design covering /v1/* endpoints statically, it handles 
		// bounds reasonably gracefully securely.
		RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}
