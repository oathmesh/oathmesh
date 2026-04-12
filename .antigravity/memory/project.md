---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
review_by: "2026-07-05"
purpose: "Persistent architecture knowledge — core abstractions, module map, tech debt"
---

# OathMesh Project Memory

## Architecture Overview

OathMesh is a micro-protocol and developer product for short-lived signed machine-call identity. The architecture follows a clear separation: the Issuer authenticates callers and mints tokens; Receivers verify tokens and authorize requests locally.

### Component Map

```
┌─────────────────────────────────────────────────────────────┐
│                     OathMesh System                         │
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────────────────┐   │
│  │  Caller   │───▶│  Issuer  │    │     Receiver         │   │
│  │ (agent,   │◀───│ Service  │    │  ┌────────────────┐  │   │
│  │  CI job,  │    │          │    │  │   Verifier     │  │   │
│  │  bot)     │    │ Mint API │    │  │   Middleware    │  │   │
│  └──────────┘    │ JWKS     │    │  └───────┬────────┘  │   │
│       │          │ Metadata │    │  ┌───────▼────────┐  │   │
│       │          └──────────┘    │  │  Policy Engine │  │   │
│       │               │          │  └───────┬────────┘  │   │
│       │          ┌────▼─────┐    │  ┌───────▼────────┐  │   │
│       └─────────▶│ Gateway  │───▶│  │   Application  │  │   │
│                  │ (opt.)   │    │  │   Code         │  │   │
│                  └──────────┘    │  └───────┬────────┘  │   │
│                                  │  ┌───────▼────────┐  │   │
│                                  │  │  Audit Logger  │  │   │
│                                  │  └────────────────┘  │   │
│                                  └──────────────────────┘   │
│                                                             │
│  ┌──────────┐                                               │
│  │   CLI    │ — mint, verify, inspect, serve, keys          │
│  └──────────┘                                               │
└─────────────────────────────────────────────────────────────┘
```

### Module Map (Planned)

| Module | Language | Purpose | Status |
|---|---|---|---|
| `/issuer` | Go | Issuer service — mint API, JWKS, metadata, key management | Not started |
| `/sdk-node` | TypeScript | Node.js verifier middleware (Express/Koa/Hono) | Not started |
| `/sdk-python` | Python | Python verifier middleware (FastAPI/Flask) | Not started |
| `/cli` | Go | CLI tool — resides in `/issuer/cmd/oathmesh/` | Not started |
| `/gateway` | Go | Reverse proxy with token verification | Not started |
| `/examples` | Mixed | Working examples (Node, Python, GitHub Actions, curl) | Not started |
| `/docs` | Markdown | Documentation source | Not started |
| `/spec` | JSON/YAML | JSON Schemas, protocol specification files | Not started |

## Core Abstractions

| Abstraction | Description | Key Interface |
|---|---|---|
| `Signer` | Signs token payloads. Implementations: local file key, KMS | `Sign(payload) → signature` |
| `Verifier` | Validates token signatures against cached JWKS | `Verify(token) → VerifiedCallerContext \| error` |
| `PolicyEvaluator` | Evaluates verified context against YAML policy rules | `Evaluate(context, request) → allow/deny + rule name` |
| `AuditEmitter` | Emits structured audit events to configured output | `Emit(event)` |
| `JWKSCache` | Fetches and caches issuer public keys with TTL-based refresh | `GetKey(issuer, kid) → publicKey \| error` |
| `TokenMinter` | Orchestrates claim construction, signing, and response | `Mint(request, callerIdentity) → OathToken` |

## Key Design Decisions

All major decisions are recorded as ADRs in `decisions/`. Summary:

| ADR | Decision | Date |
|---|---|---|
| ADR-001 | Token format: JWT/JWS with `om+jwt` type, EdDSA preferred | 2026-04-05 |
| ADR-002 | Auth: caller authenticates to issuer; receiver authorizes locally | 2026-04-05 |
| ADR-003 | Tech stack: Go issuer, TypeScript Node SDK, Python SDK | 2026-04-05 |

## Known Technical Debt

| Item | Description | Priority | Created |
|---|---|---|---|
| TD-001 | Issuer authentication methods (beyond GitHub OIDC) not formally specified | High | 2026-04-05 |
| TD-002 | `act` vs `scope` claim semantics need formal relationship definition | High | 2026-04-05 |
| TD-003 | Replay cache backing store (in-memory vs Redis vs embedded) not decided | Medium | 2026-04-05 |
| TD-004 | Error taxonomy not yet enumerated — only example format exists | Medium | 2026-04-05 |
| TD-005 | Audit event schema defined in `.antigravity/skills/data-modeling.md` but no JSON Schema yet | Medium | 2026-04-05 |
| TD-006 | No OpenAPI spec for issuer REST API | Low | 2026-04-05 |

## Known Constraints

- Token format is frozen (ADR-001) — no changes without ADR amendment
- Subject URI schemes are frozen — `svc://`, `agent://`, `job://`, `tool://`, `user://`
- Default-deny policy — cannot be changed
- Maximum TTL: 300 seconds — protocol hard limit
- Issuer must be stateless — no session or token state
