# gRPC Server Troubleshooting

## Request Path (gRPC)

```text
grpcurl/client --> metadata: authorization --> unary/stream interceptor --> handler
                                              | pass                    | fail
                                              v                         v
                                       VerifiedCallerContext         gRPC status error
```

## Quick Checks

| Check | Command |
|---|---|
| Build server | `go build -o grpc-server .` |
| Run server | `./grpc-server` |
| Verify port | `grpcurl -plaintext localhost:50051 list` |
| Call with token | `grpcurl -plaintext -H "authorization: OathMesh $TOKEN" ...` |

## Common Issues

| Issue | Cause | Fix |
|---|---|---|
| `Unauthenticated` / missing token | No authorization metadata | Add `-H "authorization: OathMesh <token>"` |
| `PermissionDenied` | Policy denied or subject revoked | Inspect policy and revocation list |
| `ResourceExhausted` | Rate limiter threshold reached | Wait/reset or raise rate limit config |
| Connection failure on `50051` | Server not listening or wrong port | Validate `OATHMESH_PORT` and process state |
