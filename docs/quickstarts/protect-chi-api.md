# Quickstart: Protect a Go chi API

**Time:** ~5 minutes

This guide adds OathMesh token verification to an existing Go chi API using the `sdk/go/middleware` package.

## Prerequisites

- Go 1.22+
- A running OathMesh issuer (or use `docker-compose up` from the repo root)
- A private key for minting tokens (`openssl genpkey -algorithm Ed25519 -out private.pem`)

## Step 1: Import the Middleware

```go
import (
    "github.com/oathmesh/oathmesh/sdk/go/middleware"
    "github.com/oathmesh/oathmesh/internal/verify"
)
```

## Step 2: Configure the Verifier

```go
cfg := &verify.VerifierConfig{
    Audience:       "https://inventory.internal",
    TrustedIssuers: []string{"https://issuer.oathmesh.dev"},
    JWKSProvider:   verify.NewJWKSCache(verify.DefaultJWKSCacheTTL),
    ReplayCache:    verify.NewMemoryReplayCache(),
}
```

## Step 3: Mount the Middleware

```go
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
            "caller": caller.Principal.Subject,
            "action": caller.Action,
        })
    })
})
```

## Step 4: Test It

```bash
# Mint a token
TOKEN=$(oathmesh mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "inventory.read" \
  --quiet)

# Call the protected endpoint
curl -H "Authorization: OathMesh $TOKEN" http://localhost:8080/inventory
```

Expected output:
```json
{"caller":"agent://repo/acme/deploy-bot","action":"inventory.read"}
```

## What Happens on Failure

| Scenario | Response |
|---|---|
| Missing token | `401 {"error":"claim_missing:token"}` |
| Wrong audience | `401 {"error":"audience_mismatch"}` |
| Expired token | `401 {"error":"token_expired"}` |
| Replayed token | `401 {"error":"replay_detected"}` |

## Next Steps

- [Protect an Express API](protect-express-api.md)
- [Run the full demo](local-demo-docker-compose.md)
