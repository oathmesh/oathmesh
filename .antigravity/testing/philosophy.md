---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "OathMesh testing philosophy — boundaries, naming, coverage, mocking strategy"
---

# Testing Philosophy

## Guiding Principle

Tests exist to provide confidence that OathMesh behaves correctly — especially in the security-critical verification and policy evaluation paths. Tests are not bureaucracy; they are the safety net that allows fast iteration.

## Test Boundaries

| Layer | What It Tests | Dependencies | Speed Target | Coverage Target |
|---|---|---|---|---|
| **Unit** | Single function logic | Mocked/stubbed | < 100ms each | ≥ 80% new code |
| **Integration** | Component interactions | In-process fakes | < 1s each | Key paths covered |
| **End-to-End** | Full mint → verify → policy → audit flow | Docker Compose | < 30s each | Golden path + error paths |
| **Benchmark** | Latency and throughput | Real components | N/A | Performance budgets met |

## Module-Specific Expectations

### Issuer (/issuer)
- **Unit**: Token minting, claim construction, key loading, config parsing
- **Integration**: HTTP handler + minter + signer end-to-end
- **Coverage**: ≥ 80% overall, 100% for signing and key rotation paths
- **Benchmark**: Token minting latency (target: < 10ms p99 excluding KMS)

### Verifier (embedded in SDKs)
- **Unit**: Each of the 14 verification steps independently
- **Integration**: Full verification chain with test JWKS server
- **Coverage**: ≥ 90% — this is security-critical code
- **Benchmark**: Verification latency (target: < 2ms p99 with cached JWKS)

### Policy Engine (embedded in SDKs)
- **Unit**: Rule matching, glob patterns, default-deny
- **Integration**: Policy loading + evaluation with real YAML files
- **Coverage**: ≥ 90% — authorization decisions must be tested
- **Benchmark**: Evaluation latency (target: < 1ms p99 for 100 rules)

### CLI
- **Unit**: Argument parsing, output formatting
- **Integration**: CLI → issuer interaction (using test server)
- **Coverage**: ≥ 70% — CLI output formatting is tested manually

### Gateway
- **Unit**: Header stripping, trust header construction, HMAC signing
- **Integration**: Full proxy flow with test upstream
- **Coverage**: ≥ 80%, 100% for trust header handling

## What Is Acceptable to Ship Untested

Explicitly:
1. CLI output formatting (colors, tables) — visual verification
2. Docker Compose example configurations — tested by running
3. Documentation prose — reviewed by Docs Agent
4. Log output formatting — verified manually in dev

Everything else requires automated tests.

## Mocking Strategy

| Component | Mock Approach |
|---|---|
| Signing keys | Generate ephemeral Ed25519 pairs in test setup |
| JWKS endpoint | `httptest.Server` (Go), `nock` (Node), `responses` (Python) |
| Time/clock | Inject clock dependency; freeze time in tests |
| KMS | Interface with in-memory fake implementation |
| Filesystem (policy files) | Temp directory with test policy files |
| HTTP requests | `httptest.Server` (Go), `supertest` (Node), `httpx.ASGITransport` (Python) |

### Anti-Patterns

- Never use production keys in tests
- Never hardcode JWT strings — mint fresh tokens in test setup
- Never mock the verification logic itself when testing — test the real verification
- Never test implementation details — test observable behavior
- Never write tests that depend on execution order

## Test Data Management

- **Golden files**: Store expected token structures in `testdata/` directories
- **Test policy files**: Create minimal, focused policy files per test scenario
- **Test fixtures**: Reusable test data shared across test files in a `testutil` package
- **Update mechanism**: `go test -update-golden` (Go), `UPDATE_GOLDEN=1 npx vitest` (Node)

## Naming Conventions

See `agents/test-agent.md` § Test Naming Conventions for language-specific conventions.

## Quality Gate Enforcement

Per `agents/test-agent.md`, a task is NOT complete until all quality gates pass. The AI must run tests and verify coverage before declaring any implementation task done.
