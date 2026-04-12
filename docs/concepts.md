# OathMesh Concepts

## Oath Token

An Oath Token is a compact, signed JWS (JSON Web Signature) that represents a single call assertion. It contains:

- **Who** is calling (`sub` — a URI like `agent://repo/acme/deploy-bot`)
- **What** they want to do (`act` — like `inventory.write`)
- **Who** they're calling (`aud` — like `https://inventory.internal`)
- **When** it was issued (`iat`) and when it expires (`exp`)
- **A unique ID** (`jti` — UUID, never reusable)

Oath Tokens use `typ: "om+jwt"` in the header to distinguish them from standard JWTs. They are always signed with `EdDSA` (Ed25519).

**Oath Tokens are not refresh tokens.** They are single-use, short-lived call credentials. When a token expires, the caller mints a new one.

## Issuer

The Issuer is an HTTP service that mints and signs Oath Tokens. It:

- Holds the Ed25519 private key (loaded from environment, never hardcoded)
- Publishes public keys via `GET /.well-known/jwks.json`
- Enforces TTL ceilings server-side (max 300 seconds, no exceptions)
- Provides `POST /v1/token` for direct minting
- Provides `POST /v1/exchange/github` for GitHub Actions OIDC token exchange

The Issuer is the **only** component that holds private key material.

## Caller

A Caller is any agent, bot, CI job, service, or tool that requests an Oath Token and presents it to a Receiver. Callers are identified by Subject URIs:

| Scheme | Use |
|---|---|
| `svc://namespace/name` | Services and microservices |
| `agent://repo/org/name` | AI agents and bots |
| `job://scheduler/name` | CI/CD jobs |
| `tool://runtime/client` | MCP-adjacent tool clients |
| `user://id` | Human delegation context only |

## Receiver

A Receiver is any service that accepts incoming requests bearing Oath Tokens. The Receiver runs the 14-step verification pipeline (either via the Go SDK middleware, a polyglot SDK, or the OathMesh Gateway) and gets a `VerifiedCallerContext` on success.

**The Receiver makes authorization decisions.** OathMesh tells the Receiver *who* is calling; the Receiver decides *what they can do*.

## VerifiedCallerContext

The `VerifiedCallerContext` is the structured output of a successful token verification. It contains:

```
Principal:
  Issuer:   https://issuer.oathmesh.dev
  Subject:  agent://repo/acme/deploy-bot

Action:     inventory.write
Scope:      [inventory.read, inventory.write]
Reason:     "sync catalog after deploy"
TokenID:    550e8400-e29b-41d4-a716-446655440000
IssuedAt:   2026-04-12T14:30:00Z
ExpiresAt:  2026-04-12T14:32:00Z
Env:        prod

Source:
  Type:     github_actions
  Repo:     acme/storefront
  Workflow: deploy.yml
  RunID:    123456
  SHA:      abc123def456
```

This context is what receivers use to make authorization decisions. It is never the raw token—it is a parsed, verified, trusted representation of the caller's identity.

## Pkl Policy

OathMesh uses [Apple Pkl](https://pkl-lang.org/) for policy files. Pkl provides:

1. **Schema validation at load time** — malformed policies are caught before any request is processed, not at runtime.
2. **Type safety** — the policy schema (`policy/schema.pkl`) defines exact types for rules, match criteria, and source provenance. Invalid values are compile-time errors.

Pkl is used in two places:

- **`policy/schema.pkl`** — defines the policy rule schema (match criteria, allow/deny rules)
- **`internal/config/issuer.pkl`** — defines the issuer configuration schema (TTL, rate limiting, audit, replay cache)

Policy rules are evaluated in order. First match wins. The last rule **must** be a default deny:

```pkl
new {
  name = "default"
  allow = false
}
```

Policies hot-reload on file change via `fsnotify` — zero downtime, atomic swap.

## Gateway Mode

The OathMesh Gateway is an optional reverse proxy mode (`oathmesh serve --gateway`) that sits in front of upstream services. It:

1. Intercepts incoming requests
2. Extracts the Oath Token from the `Authorization` header
3. Runs the full 14-step verification pipeline
4. If denied: returns 401, does not forward
5. If allowed: **strips the raw token entirely**, injects `X-OathMesh-*` context headers, and forwards

The upstream service never sees the raw token. It receives verified identity through headers:

```
X-OathMesh-Subject:  agent://repo/acme/deploy-bot
X-OathMesh-Action:   inventory.write
X-OathMesh-Token-Id: <jti>
X-OathMesh-Issuer:   https://issuer.oathmesh.dev
X-OathMesh-Env:      prod
```

## Audit Events

Every verification attempt—whether allowed or denied—emits a structured NDJSON audit event. Events are never conditional. They include:

- `event`: always `"oathmesh.verify"`
- `outcome`: `"allow"` or `"deny"`
- `reason`: which policy rule matched (or the error code)
- `jti`, `sub`, `aud`, `act`, `iss`, `env`: claim summary
- `src`: source provenance if present
- `timestamp` and `request_id`

**Rule:** Never log the full token string. Log `jti` + claim summary only. Never log private key material.
