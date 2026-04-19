# OathMesh gRPC Server Example

This example demonstrates how to integrate OathMesh token authentication and authorization with gRPC services.

## Overview

The gRPC server example implements a simple `UserService` with the following features:

- **Token Verification**: All incoming gRPC requests are validated using OathMesh tokens
- **Rate Limiting**: Per-subject rate limiting (configurable limit per minute)
- **Metadata Extraction**: Verified claims are injected into the gRPC context for handler use
- **Error Handling**: Proper gRPC error codes for different failure scenarios
- **Monitoring**: Prometheus metrics for request counting and latency

## Prerequisites

- Go 1.26.2+
- Protocol Buffers compiler (protoc) with Go code generation plugins
- OathMesh token issuer running (for token generation)

## Quick Start

### 1. Build the Example

```bash
cd examples/grpc-server
go build -o grpc-server .
```

### 2. Run the Example

```bash
export OATHMESH_AUDIENCE="grpc-example-service"
export OATHMESH_TRUSTED_ISSUERS="https://issuer.example.com"
export OATHMESH_PORT="50051"

./grpc-server
```

The server will start listening on `localhost:50051`.

### 3. Test with a Valid Token

First, mint a token from the OathMesh issuer:

```bash
# This example assumes you have an OathMesh issuer running
# and can generate tokens

SUBJECT="svc://my-service"
ISSUER="https://issuer.example.com"
AUDIENCE="grpc-example-service"

# Generate a token (use your issuer's API)
TOKEN=$(curl -s https://issuer.example.com/v1/mint \
  -H "Content-Type: application/json" \
  -d "{
    \"sub\": \"$SUBJECT\",
    \"iss\": \"$ISSUER\",
    \"aud\": \"$AUDIENCE\",
    \"act\": \"read\"
  }" | jq -r '.token')
```

Then test the gRPC service using `grpcurl`:

```bash
# Health check
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  localhost:50051 grpcserver.UserService.Health

# Get user
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"user_id": "123"}' \
  localhost:50051 grpcserver.UserService.GetUser
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OATHMESH_AUDIENCE` | `grpc-example-service` | Expected audience in tokens |
| `OATHMESH_TRUSTED_ISSUERS` | `https://issuer.example.com` | Comma-separated list of trusted issuer URLs |
| `OATHMESH_POLICY_PATH` | (empty) | Path to Pkl policy file for authorization rules |
| `OATHMESH_PORT` | `50051` | Port for gRPC server |

### Example with Policy

```bash
export OATHMESH_POLICY_PATH="/path/to/policy.pkl"
./grpc-server
```

## Implementation Details

### Token Authentication

Tokens are extracted from gRPC metadata with the `authorization` header:

```
authorization: Bearer <token>
// or
authorization: OathMesh <token>
```

The middleware automatically:
1. Extracts the token from metadata
2. Verifies the token signature and claims
3. Checks token expiry and audience
4. Applies rate limiting per subject
5. Injects verified claims into the request context

### Accessing Verified Claims in Handlers

In your gRPC service implementation, access the verified caller context:

```go
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // Extract verified claims
    vcc := middleware.VerifiedCallerFrom(ctx)
    if vcc == nil {
        return nil, status.Error(codes.Unauthenticated, "missing verified claims")
    }
    
    // Use claims in your handler
    subject := vcc.Principal.Subject
    issuer := vcc.Principal.Issuer
    
    // ... handler logic
}
```

### Error Handling

The middleware maps OathMesh errors to gRPC error codes:

| OathMesh Error | gRPC Code | HTTP Equivalent |
|---|---|---|
| Missing/Invalid Token | `Unauthenticated` (16) | 401 |
| Expired Token | `Unauthenticated` (16) | 401 |
| Signature Invalid | `Unauthenticated` (16) | 401 |
| Policy Denied | `PermissionDenied` (7) | 403 |
| Subject Revoked | `PermissionDenied` (7) | 403 |
| Rate Limit Exceeded | `ResourceExhausted` (8) | 429 |
| Internal Error | `Internal` (13) | 500 |

### Rate Limiting

The example uses a per-subject rate limiter with configurable limits:

```go
// 1000 tokens per minute per subject (default)
rateLimiter := middleware.NewSimpleRateLimiter(1000)

// Custom limit
rateLimiter := middleware.NewSimpleRateLimiter(5000)
```

Rate limits are enforced per token subject (the `sub` claim):
- Each subject has an independent rate limit
- Requests from different subjects don't affect each other
- Limits reset every minute

### Metrics

The gRPC middleware exports Prometheus metrics:

```
# Request count by method and status
oathmesh_grpc_requests_total{method="GetUser",status="OK"} 42
oathmesh_grpc_requests_total{method="GetUser",status="Unauthenticated"} 3

# Request duration by method
oathmesh_grpc_request_duration_seconds_bucket{method="GetUser",le="0.1"} 40
oathmesh_grpc_request_duration_seconds_bucket{method="GetUser",le="0.5"} 42
```

Access metrics via:
```bash
curl http://localhost:8080/metrics | grep oathmesh_grpc
```

## Docker Deployment

### Build the Docker Image

```bash
docker build -f examples/grpc-server/Dockerfile -t oathmesh-grpc-server:latest .
```

### Run in Docker

```bash
docker run -d \
  -p 50051:50051 \
  -e OATHMESH_AUDIENCE="grpc-example-service" \
  -e OATHMESH_TRUSTED_ISSUERS="https://issuer.example.com" \
  oathmesh-grpc-server:latest
```

## Testing

### Unit Tests

Test the gRPC middleware directly:

```bash
go test ./internal/middleware/... -v -cover
```

### Integration Tests

Test with a real gRPC client:

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 list grpcserver.UserService
```

## Troubleshooting

See the dedicated guide: [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)


### Issue: "missing authorization token"

**Cause**: Token not provided in request

**Solution**: Add the `authorization` header:
```bash
grpcurl -H "authorization: Bearer <token>" ...
```

### Issue: "token expired"

**Cause**: Token has expired

**Solution**: Generate a fresh token from the issuer

### Issue: "policy denied"

**Cause**: Token failed policy evaluation

**Solution**: Check the Pkl policy file and token claims

### Issue: "rate limit exceeded"

**Cause**: Subject has exceeded the request limit

**Solution**: Wait a minute for the rate limit to reset, or increase `RateLimitPerMinute`

## Next Steps

1. **Implement Service Methods**: Add actual business logic to the `UserService` methods
2. **Add Policy Enforcement**: Create a Pkl policy file to define authorization rules
3. **Enable TLS**: Use `grpc.Creds(credentials.NewServerTLSFromFile(...))` for production
4. **Add Logging**: Integrate structured logging via the handler context
5. **Scale**: Deploy multiple instances behind a load balancer

## References

- [OathMesh Documentation](https://github.com/oathmesh/oathmesh)
- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
