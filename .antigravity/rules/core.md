---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Non-negotiable behavioral and architectural rules for all OathMesh AI work"
---

# OathMesh Core Rules

These rules are absolute. They override all other guidance, including skill suggestions, agent recommendations, and persona preferences. When in doubt, these rules win.

## Priority Hierarchy

When rules conflict, resolve in this fixed order:

1. **Security** — `security/redlines.md`, `rules/security-redlines.md`
2. **Architecture** — `decisions/`, `rules/core.md` (this file)
3. **Performance** — performance budgets in this file
4. **Style** — `rules/coding-standards.md`, `personas/voice.md`

Full resolution protocol: `rules/conflict-resolution.md`

## Architectural Constraints

### AC-1: Token Format Is Frozen

The Oath Token format is signed compact JWS/JWT with `typ: om+jwt`. This is decided and will not change. See `decisions/ADR-001-token-format.md`.

Do not:
- Suggest alternative token formats (PASETO, Macaroons, custom binary)
- Propose removing the `om+jwt` type header
- Recommend non-standard JWT libraries

### AC-2: Authentication vs. Authorization Separation

OathMesh authenticates callers. Receivers authorize requests. These responsibilities must never be mixed.

Do not:
- Add authorization logic to the issuer
- Add authentication logic to the policy engine
- Make the issuer aware of receiver-side policy rules

### AC-3: Receiver-Local Authorization

Policy evaluation happens at the receiver, not at a central policy server. Policy files are local to the receiver.

Do not:
- Design a central policy service that receivers call on every request
- Add network-dependent authorization checks to the verification path

### AC-4: Stateless Issuer

The issuer service must be stateless and horizontally scalable. It holds signing keys (via KMS or local config) and issues tokens. It does not store session state, token state, or caller state.

### AC-5: Monorepo Structure

OathMesh uses a monorepo with these top-level directories:

```
/issuer       — Go issuer service
/sdk-node     — Node.js verifier middleware (npm package)
/sdk-python   — Python verifier middleware (pip package)
/cli          — CLI tool (Go, same binary as issuer or standalone)
/gateway      — Gateway/reverse-proxy mode (Go)
/examples     — Working examples (Node, Python, GitHub Actions, curl)
/docs         — Documentation source
/spec         — JSON Schemas, protocol spec files
```

See `decisions/ADR-003-tech-stack.md` for language rationale.

## Behavioral Rules

### BR-1: Check Before Changing Architecture

Before proposing any change that affects component boundaries, data flow, or public API surface:

1. Read all active ADRs in `decisions/`
2. Check `context/current-phase.md` for off-limits zones
3. If the change contradicts an ADR, propose an ADR amendment first — do not just make the change

### BR-2: Redline Enforcement

Before executing any action listed in `security/redlines.md`, stop and request explicit human approval. Never proceed on assumption.

### BR-3: Ambiguity Protocol

When context is missing or signals conflict:

1. State what is known and what is unknown
2. Identify which `.antigravity/` file should contain the answer
3. If the file exists but doesn't cover this case, propose an update to that file
4. If the file doesn't exist, flag the gap and suggest creating it
5. If the uncertainty involves security, data integrity, or irreversible actions, ask the human before proceeding
6. For low-stakes uncertainty (naming, formatting, minor structure), make a reasonable choice, document the reasoning, and flag it for review

### BR-4: Terminology Enforcement

Use only the terms defined in `context/glossary.md`. When writing code, comments, docs, error messages, or conversation — use the canonical term. If a new term is needed, propose adding it to the glossary.

### BR-5: Specification Fidelity

The source-of-truth for OathMesh behavior is `oathmesh.txt` in the project root. When implementation details are ambiguous, consult `oathmesh.txt` first. When `oathmesh.txt` and an ADR conflict, the ADR takes precedence (ADRs represent decisions made after the initial spec).

## Performance Budgets

These are target budgets, not hard limits. Exceeding them requires justification.

| Metric | Budget | Applies To |
|---|---|---|
| Token minting latency (p99) | < 10ms excluding KMS call | Issuer service |
| Token verification latency (p99) | < 2ms with cached JWKS | Verifier middleware |
| Policy evaluation latency (p99) | < 1ms for 100-rule policy file | Policy engine |
| JWKS refresh interval | 5 minutes default, configurable | Verifier middleware |
| Token size | < 2KB for Core Profile | Token format |
| Gateway added latency (p99) | < 5ms | Gateway mode |
| CLI command startup | < 200ms | CLI tool |
| Docker image size | < 30MB | Issuer, gateway, CLI |

## Dependency Rules

### DR-1: Minimal Dependencies

Prefer standard library over third-party libraries, especially for:
- JSON handling
- HTTP server/client
- Cryptographic operations (use `crypto/ed25519` in Go, not a wrapper)

### DR-2: No Vendored Crypto

Never implement custom cryptographic algorithms. Use well-audited, standard-library or CNCF-grade crypto packages only.

### DR-3: Transitive Dependency Awareness

Before adding a dependency, check:
- Does it pull in a large transitive tree?
- Is it actively maintained?
- Does it have known CVEs?
- Can the same result be achieved with the standard library?

Document the justification for any dependency with more than 3 transitive dependencies.

## Documentation Rules

### DOC-1: Every Public API Has Documentation

Every exported function, endpoint, CLI command, and configuration option must have documentation before it ships. Not after. Not "we'll add docs later."

### DOC-2: Error Messages Are Documentation

Every error returned by OathMesh must explain: what failed, why it failed, and what to fix. See `oathmesh.txt` section 24.3 for the error experience standard.

Example:
```
audience_mismatch: token was minted for https://billing.internal but received by https://inventory.internal
```

### DOC-3: Examples Are Tests

Every code example in documentation must be extracted from or validated against actual running code. Stale examples are bugs.
