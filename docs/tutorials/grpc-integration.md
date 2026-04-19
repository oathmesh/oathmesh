← [Back to Index](../INDEX.md)

# Tutorial: gRPC Integration (Go)

This tutorial covers the Phase 2 gRPC surfaces and points to the exact repo files used in production wiring.

## What to use

- SDK wrappers: `sdk/go/middleware/grpc.go`
- Core interceptors: `internal/middleware/grpc_unary.go`, `internal/middleware/grpc_stream.go`
- Example wiring: `examples/grpc-server/main.go`

## 1) Verify unary + stream protected call paths

From repo root:

```bash
go test ./internal/middleware -run "TestUnaryInterceptor_ValidToken_ContextInjectionPath|TestStreamInterceptor_ValidToken_ContextInjectionPath|TestUnaryInterceptor_MissingToken" -v
```

Expected: all selected tests `PASS` (token extraction, verification, context injection, auth deny path).

## 2) Run the example server wiring

```bash
OATHMESH_AUDIENCE=grpc-example-service \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
go run ./examples/grpc-server
```

Expected logs include:

- `gRPC server listening on port 50051 with OathMesh authentication enabled`
- configured audience + trusted issuers

## 3) Register middleware in your own gRPC server

```go
opts := []grpc.ServerOption{
  grpc.UnaryInterceptor(middleware.OathMeshUnaryInterceptor(verifierCfg, nil)),
  grpc.StreamInterceptor(middleware.OathMeshStreamInterceptor(verifierCfg, nil)),
}
server := grpc.NewServer(opts...)
```

In handlers:

```go
vcc := middleware.VerifiedCallerFrom(ctx)
```

## Notes

- The example server currently demonstrates middleware wiring; add your own protobuf service registration to issue live `grpcurl` method calls.
- Authorization metadata supports both `authorization: OathMesh <token>` and `authorization: Bearer <token>`.
