---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
---

# ADR-002: Authentication Strategy — Caller Auth to Issuer, Receiver-Local Authorization

**Status:** Accepted

## Context

OathMesh's core architecture requires two distinct trust decisions:
1. How does the issuer know who the caller is? (Authentication)
2. How does the receiver decide if the caller is allowed? (Authorization)

These responsibilities must be cleanly separated to keep the protocol simple and avoid centralized bottlenecks.

Reference: `oathmesh.txt` sections 7.2, 8.2, 13.1–13.4

## Options Considered

### Option A: Centralized Auth + Centralized Authz

- Issuer authenticates callers AND evaluates authorization
- Receiver blindly trusts tokens from the issuer
- Cons: single point of failure, latency bottleneck, issuer must know every receiver's policy
- Cons: violates "receiver authorizes the request" doctrine

### Option B: Caller Auth to Issuer + Receiver-Local Authorization (CHOSEN)

- Issuer authenticates the caller and mints a token with identity + context
- Receiver verifies the token locally (signature + claims) and evaluates its own policy
- Pros: no network dependency for authorization decisions
- Pros: receivers own their own security policy
- Pros: issuer stays stateless and simple
- Cons: policy files must be distributed to each receiver

### Option C: Federated Authorization Service

- Separate policy service that receivers call for authorization decisions
- Pros: centralized policy management
- Cons: adds network dependency to every request path
- Cons: single point of failure for authorization
- Cons: adds latency to every verified request

## Decision

We will use **Option B: Caller authenticates to the issuer; the receiver authorizes locally.**

This maps directly to the OathMesh doctrine: "OathMesh authenticates the caller. The receiver authorizes the request."

### Caller-to-Issuer Authentication Methods

The issuer supports multiple bootstrap authentication methods, with the caller's identity derived from the bootstrap credential — never self-asserted:

| Method | Bootstrap Credential | Subject Derivation | Priority |
|---|---|---|---|
| GitHub Actions OIDC | GitHub-signed OIDC JWT | `job://github-actions/{owner}/{repo}/{workflow}` | Primary (golden path) |
| Kubernetes ServiceAccount | K8s SA JWT | `svc://{namespace}/{service-account}` | Secondary |
| Client Credentials | Pre-shared ID + secret | Configured `sub` value | Dev mode only |

The issuer verifies the bootstrap credential independently (e.g., validates GitHub OIDC against GitHub's JWKS, validates K8s SA against cluster OIDC discovery).

### Receiver-Local Authorization

Each receiver maintains its own YAML policy file. Policy evaluation is:
- Local — no network calls during policy evaluation
- Default-deny — if no rule matches, the request is denied
- First-match — rules are evaluated top-to-bottom, first match wins
- File-based — policy is a YAML file, not a database or API

Policy files support matching on: `sub`, `act`, `scope`, `src.*`, `env`.

## Consequences

### Positive
- The issuer is simple and stateless — easy to scale horizontally
- Receivers are self-contained — no external dependency for authorization
- Policy changes are local to the receiver — no coordination needed
- Verification latency is sub-millisecond (no network calls)
- Clear security boundary: issuer knows identity; receiver knows policy

### Negative
- Policy files must be distributed to receivers (file sync, config management, or baked into deployment)
- No central visibility into all policies (addressed in Phase 3 with managed policy registry)
- Receivers must handle JWKS caching independently

### Risks
- Stale policy files: mitigated by policy file versioning and reload-on-change
- JWKS cache staleness: mitigated by TTL-based refresh and force-refresh on kid miss

## References

- `oathmesh.txt` sections 7.2 (core flow), 8.2 (protocol doctrine), 13.1–13.4 (authorization model)
- `skills/auth.md` (authentication implementation details)
- `skills/data-modeling.md` (policy file schema)
- ADR-001 (token format — the token carries auth context, not authz decisions)
