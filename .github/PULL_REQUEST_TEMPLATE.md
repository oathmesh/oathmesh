## Description

What does this PR do? Why is it needed?

## Changes

- 

## Testing

- [ ] `go test -race ./...` passes
- [ ] `go vet ./...` passes
- [ ] New tests added for new functionality
- [ ] Existing tests still pass

## Security Checklist

If this PR touches `internal/sign/`, `internal/verify/`, or any crypto/token code:

- [ ] Private key material is never logged, never returned in any HTTP response
- [ ] Full Oath Token strings are never logged — `jti` + claim summary only
- [ ] `alg: "none"` is rejected at Step 02 before any other step runs
- [ ] Audience matching is exact string comparison (no globs, no prefix/suffix)
- [ ] TTL is enforced server-side, maximum 300 seconds, callers cannot influence `exp`
- [ ] `jti` is generated via `uuid.New()` — cryptographically random
- [ ] JWKS fetch uses a dedicated `http.Client` with 5-second timeout (not `http.DefaultClient`)
- [ ] Replay cache operations are thread-safe (`sync.RWMutex` for memory, `SET NX` for Redis)
- [ ] Audit events are emitted on every verification (allow AND deny, no exceptions)

## Documentation

- [ ] `docs/` updated if public-facing behavior changed
- [ ] `ARCHITECTURE.md` updated if package structure changed
- [ ] New ADR written if protocol, schema, or architecture changed

## Breaking Changes

List any breaking changes and migration instructions:

- None
