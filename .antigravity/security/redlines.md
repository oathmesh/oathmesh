---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Canonical copy of security redlines — synced with rules/security-redlines.md"
---

# Security Redlines

This is an alias of `rules/security-redlines.md`. Both files must remain in sync. The authoritative copy is `rules/security-redlines.md`.

See [rules/security-redlines.md](../rules/security-redlines.md) for the complete list of irreversible actions requiring human approval.

## Quick Reference

Actions that ALWAYS require human approval:

1. **Cryptographic operations**: key generation, key rotation, algorithm changes, JWKS modifications
2. **Auth/authz changes**: issuer auth methods, claim changes, TTL changes, policy logic, trusted issuers
3. **Data operations**: schema migrations, audit log deletion, replay cache clearing
4. **Deployment**: production deploys, Docker base image changes, CI/CD secret changes, TLS config
5. **Configuration**: `.env` file changes, environment URLs, `.gitignore` security patterns
6. **Protocol changes**: token type header, subject URI schemes, auth scheme, metadata endpoint path (all FROZEN)

When in doubt, ask first. "I assumed it was fine" is never acceptable.
