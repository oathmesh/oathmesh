---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Actions that ALWAYS require human approval — no exceptions"
---

# OathMesh Security Redlines

These are irreversible or high-risk actions that must never be performed without explicit human approval. The AI must halt and request confirmation before proceeding. "I assumed it was fine" is never an acceptable justification.

This file is the authoritative copy. `security/redlines.md` is an alias that must remain in sync.

## Category 1: Cryptographic Operations

| Action | Risk | Gate |
|---|---|---|
| Generating new signing key pairs | Wrong algorithm or key size compromises all tokens | Human approves algorithm and storage method |
| Rotating signing keys | Premature rotation without overlapping publication breaks verification | Human approves rotation plan and timing |
| Changing supported algorithms in issuer metadata | Removing an algorithm breaks existing verifiers | Human approves the change and migration plan |
| Modifying JWKS output | Incorrect JWKS breaks all token verification | Human reviews the JWKS structure before deployment |

## Category 2: Authentication & Authorization

| Action | Risk | Gate |
|---|---|---|
| Changing issuer authentication methods | Wrong method weakens the trust root | Human approves |
| Modifying token claims (adding/removing required claims) | Breaks protocol compatibility | Human approves + ADR required |
| Changing default TTL values | Too long increases replay risk; too short breaks workflows | Human approves |
| Modifying the policy evaluation logic | Wrong change silently allows or denies requests | Human reviews diff + test results |
| Adding or removing trusted issuers in any config | Wrong issuer relationship breaks security model | Human approves each issuer change |

## Category 3: Data & State

| Action | Risk | Gate |
|---|---|---|
| Database schema migrations | Irreversible data structure changes | Human approves migration SQL/script |
| Deleting audit logs or audit data | Loss of compliance evidence | Human approves with written justification |
| Clearing the replay cache | Enables replay of recently-used tokens | Human approves with explicit risk acknowledgment |
| Modifying audit event schema | Breaks downstream log consumers | Human approves + test with existing consumers |

## Category 4: Deployment & Infrastructure

| Action | Risk | Gate |
|---|---|---|
| Deploying to production | Any unreviewed change reaches production | Human approves deployment |
| Changing Docker base images | Supply chain risk | Human approves the specific image and tag |
| Modifying CI/CD pipeline secrets | Wrong secret exposure compromises production | Human manages secrets directly — AI never touches |
| Changing network-facing ports or TLS configuration | Exposure risk | Human approves |

## Category 5: Configuration & Environment

| Action | Risk | Gate |
|---|---|---|
| Creating or modifying `.env` files | Risk of accidental secret commit | Human manages `.env` files — AI advises but never writes |
| Changing environment-specific URLs (issuer, JWKS, audience) | Wrong URL breaks verification chain | Human approves per-environment |
| Modifying `.gitignore` to remove security-relevant patterns | Secrets could be committed | Human approves every `.gitignore` change |

## Category 6: Protocol Changes

| Action | Risk | Gate |
|---|---|---|
| Changing the Oath Token type header (`om+jwt`) | Breaks all existing verifiers | Frozen — requires ADR amendment + founder approval |
| Changing subject URI scheme format | Breaks all existing policies | Frozen — requires ADR amendment + founder approval |
| Changing the HTTP authorization scheme (`Authorization: Oathmesh`) | Breaks all existing callers | Frozen — requires ADR amendment + founder approval |
| Changing the metadata endpoint path (`/.well-known/oathmesh-issuer`) | Breaks all existing verifier discovery | Frozen — requires ADR amendment + founder approval |

## Enforcement

- The AI must check this file before executing any action in the categories above
- If uncertain whether an action falls under a redline, treat it as if it does — ask first
- Violations of redlines should be logged in `feedback/session-log-template.md` with a post-mortem entry
- This file has the highest priority in the entire configuration system — nothing overrides it
