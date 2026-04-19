---
title: OathMesh Go Middleware SDK
description: Add OathMesh token verification to Go HTTP and gRPC services with production-ready configuration and troubleshooting guidance.
keywords: [oathmesh, go, middleware, grpc, jwt, jwks, security]
audience: [developers, platform-engineers, security-engineers]
---

← [Docs Index](../../../docs/INDEX.md)

# OathMesh Go Middleware SDK

Middleware and interceptors for verifying OathMesh tokens in Go services.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [HTTP Middleware](#http-middleware)
- [gRPC Interceptors](#grpc-interceptors)
- [Config Guidance](#config-guidance)
- [Error Handling](#error-handling)
- [Troubleshooting](#troubleshooting)
- [Security Notes](#security-notes)
- [Production Tips](#production-tips)
- [Related Docs](#related-docs)

## Installation

```bash
go get github.com/oathmesh/oathmesh/sdk/go/middleware
```

**Requirements:** Go 1.22+

## Quick Start

```go
cfg := &verify.VerifierConfig{
    Audience:       "https://inventory.internal",
    TrustedIssuers: []string{"https://issuer.oathmesh.tech"},
    JWKSProvider:   verify.NewJWKSCache(verify.DefaultJWKSCacheTTL, nil),
    ReplayCache:    verify.NewMemoryReplayCache(),
}

r := chi.NewRouter()
r.Use(middleware.OathMeshMiddleware(cfg))
```

## HTTP Middleware

Minimal `chi` setup:

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/oathmesh/oathmesh/internal/verify"
    "github.com/oathmesh/oathmesh/sdk/go/middleware"
)

func main() {
    cfg := &verify.VerifierConfig{
        Audience:       "https://inventory.internal",
        TrustedIssuers: []string{"https://issuer.oathmesh.tech"},
        JWKSProvider:   verify.NewJWKSCache(verify.DefaultJWKSCacheTTL, nil),
        ReplayCache:    verify.NewMemoryReplayCache(),
    }

    r := chi.NewRouter()
    r.Use(middleware.OathMeshMiddleware(cfg))
    r.Get("/inventory", func(w http.ResponseWriter, r *http.Request) {
        caller := middleware.CallerFrom(r.Context())
        if caller == nil {
            http.Error(w, "caller context missing", http.StatusInternalServerError)
            return
        }
        _ = json.NewEncoder(w).Encode(map[string]any{"subject": caller.Principal.Subject})
    })

    _ = http.ListenAndServe(":8080", r)
}
```

Also works with stdlib `net/http`:

```go
mux := http.NewServeMux()
mux.Handle("/api/", middleware.OathMeshMiddleware(cfg)(yourHandler))
```

## gRPC Interceptors

Minimal server setup:

```go
package main

import (
    "google.golang.org/grpc"

    "github.com/oathmesh/oathmesh/internal/verify"
    sdkmiddleware "github.com/oathmesh/oathmesh/sdk/go/middleware"
)

func newServer() *grpc.Server {
    cfg := &verify.VerifierConfig{
        Audience:       "grpc-inventory",
        TrustedIssuers: []string{"https://issuer.oathmesh.tech"},
        JWKSProvider:   verify.NewJWKSCache(verify.DefaultJWKSCacheTTL, nil),
        ReplayCache:    verify.NewMemoryReplayCache(),
    }

    return grpc.NewServer(
        grpc.UnaryInterceptor(sdkmiddleware.OathMeshUnaryInterceptor(cfg, nil)), // nil => default limiter
        grpc.StreamInterceptor(sdkmiddleware.OathMeshStreamInterceptor(cfg, nil)),
    )
}
```

Send tokens through gRPC metadata:

```text
authorization: OathMesh <token>
```

(`Bearer <token>` is also accepted.)

## Config Guidance

- **Audience:** Must exactly match token `aud`.
- **TrustedIssuers:** Use explicit issuer URLs; avoid broad or inferred trust.
- **JWKSProvider:** Use `verify.NewJWKSCache(...)` for normal deployments.
- **ReplayCache:** Use at least `verify.NewMemoryReplayCache()`; use shared backing store for multi-instance deployments.
- **PolicyEvaluator:** Add policy enforcement for authZ decisions.
- **RevocationList:** Enable subject revocation checks for rapid access withdrawal.
- **AuditSink:** Emit verification outcomes for incident analysis and compliance evidence.

## Error Handling

- HTTP middleware returns `401` with structured JSON (`error`, `message`, `fix`, `request_id`).
- Middleware strips `Authorization` before calling downstream handlers.
- Always nil-check `middleware.CallerFrom(ctx)` in handlers.
- gRPC interceptors map verification failures to standard gRPC codes (`Unauthenticated`, `PermissionDenied`, `ResourceExhausted`, `Internal`).

## Troubleshooting

- **Missing caller context:** Ensure middleware/interceptors run before protected handlers; nil-check extracted caller.
- **Invalid audience:** Match receiver `Audience` to minted token `aud` exactly.
- **JWKS fetch issues:** Verify issuer reachability and `/.well-known/jwks.json` availability; check timeout/network/TLS.
- **Replay/revocation problems:** Confirm replay cache durability and revocation backend health/latency.

See the global runbook: [Troubleshooting Guide](../../../docs/TROUBLESHOOTING.md).

## Security Notes

- Prefer `Authorization: OathMesh <token>` and short-lived tokens.
- Keep trusted issuer lists explicit and minimal.
- Do not log raw tokens.
- Keep clocks synchronized (NTP) to avoid false expiry/skew failures.

## Production Tips

- Deploy issuer and trust infrastructure using [Deployment Guides](../../../docs/deployment/OVERVIEW.md).
- Use [Production Checklist](../../../docs/operations/production-checklist.md) before go-live.
- Implement [Key Rotation](../../../docs/operations/key-rotation.md) and [Incident Response](../../../docs/operations/incident-response.md).
- For on-call readiness, use the [Operations Runbook](../../../docs/operations/on-call-runbook.md).
- For regulated environments, review the [Enterprise Guide](../../../docs/enterprise/README.md).

## Related Docs

- [Getting Started](../../../docs/GETTING_STARTED.md)
- [Troubleshooting](../../../docs/TROUBLESHOOTING.md)
- [Deployment Docs](../../../docs/deployment/OVERVIEW.md)
- [Operations Docs](../../../docs/operations/on-call-runbook.md)
- [Enterprise Docs](../../../docs/enterprise/README.md)
- [Community & Support](../../../docs/COMMUNITY.md)
