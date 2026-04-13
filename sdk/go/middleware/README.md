# OathMesh Go SDK

<p align="center">
  <img src="../../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  <b>Middleware for chi, stdlib net/http, and any Go HTTP framework.</b>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/oathmesh/oathmesh/sdk/go/middleware">
    <img src="https://img.shields.io/badge/Go-reference-blue" alt="Go Reference">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml">
    <img src="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/oathmesh/oathmesh" alt="License">
  </a>
</p>

---

## Installation

```bash
go get github.com/oathmesh/oathmesh/sdk/go/middleware
```

**Requirements:** Go 1.22+

---

## Quick Start

```go
cfg := &verify.VerifierConfig{
    Audience:       "https://inventory.internal",
    TrustedIssuers: []string{"https://issuer.oathmesh.dev"},
}

r := chi.NewRouter()
r.Use(middleware.OathMeshMiddleware(cfg))
r.Get("/inventory", handler)
```

---

## chi Middleware

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/oathmesh/oathmesh/sdk/go/middleware"
    "github.com/oathmesh/oathmesh/internal/verify"
)

func main() {
    cfg := &verify.VerifierConfig{
        Audience:       "https://inventory.internal",
        TrustedIssuers: []string{"https://issuer.oathmesh.dev"},
        JWKSProvider:   verify.NewJWKSCache(verify.DefaultJWKSCacheTTL),
        ReplayCache:    verify.NewMemoryReplayCache(),
    }

    r := chi.NewRouter()

    // Public routes
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.OathMeshMiddleware(cfg))

        r.Get("/inventory", func(w http.ResponseWriter, r *http.Request) {
            caller := middleware.CallerFrom(r.Context())
            if caller == nil {
                http.Error(w, "caller context missing", 500)
                return
            }
            json.NewEncoder(w).Encode(map[string]any{
                "subject": caller.Principal.Subject,
                "action":  caller.Action,
            })
        })
    })

    log.Fatal(http.ListenAndServe(":8080", r))
}
```

## stdlib `net/http`

The middleware works with any `http.Handler`:

```go
mux := http.NewServeMux()
mux.Handle("/api/", middleware.OathMeshMiddleware(cfg)(yourHandler))
```

## API

### `OathMeshMiddleware(cfg) func(http.Handler) http.Handler`

Returns a standard Go middleware. On success, injects `VerifiedCallerContext` into the request context. On failure, returns 401 with a structured JSON error body.

The middleware also **strips the Authorization header** before passing the request to the next handler — the upstream never sees the raw token.

### `CallerFrom(ctx) *core.VerifiedCallerContext`

Extracts the verified caller from the request context. Returns `nil` if the middleware didn't run or verification failed. **Always check for nil.**

```go
caller := middleware.CallerFrom(r.Context())
if caller == nil {
    // handle missing context
}
```

## Error Responses

```json
{
  "error": "audience_mismatch",
  "message": "token audience does not match",
  "fix": "mint with --aud https://inventory.internal",
  "request_id": "req-abc-123"
}
```

The `request_id` is read from the `X-Request-Id` header if present.
