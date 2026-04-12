---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "What the AI must never log, commit, or expose"
---

# Secret Policy

## Classification

| Classification | Description | Examples |
|---|---|---|
| **CRITICAL** | Compromise grants full system access | Issuer signing private keys, KMS credentials, gateway HMAC keys |
| **HIGH** | Compromise grants significant access | Client credentials (dev mode), database connection strings, CI/CD pipeline secrets |
| **MEDIUM** | Compromise grants limited access | API keys for non-critical services, search API keys |
| **LOW** | Informational but not directly exploitable | Internal URLs, environment names, team structure |

## Rules

### R-1: Never Commit Secrets

No secret of any classification level may ever be committed to version control. This includes:
- Private keys in any format (PEM, JWK, PKCS#8, raw bytes)
- Client secrets or passwords
- API keys (even for non-critical services)
- Connection strings with credentials
- HMAC keys
- Full Oath Tokens (even expired ones — they reveal claim structure)

### R-2: .env File Handling

- `.env` files must be listed in `.gitignore`
- The AI may create `.env.example` files with placeholder values
- The AI must NEVER write or modify actual `.env` files (human responsibility per `rules/security-redlines.md`)
- Placeholder format: `OATHMESH_SIGNING_KEY=<your-private-key-path-here>`

### R-3: Logging Policy

| Data | Log Level | Format |
|---|---|---|
| Full Oath Token | **NEVER** | — |
| Token claims summary | INFO | `{sub, aud, act, jti, exp}` — omit `src` details in production |
| Private keys | **NEVER** | — |
| Public keys (JWK) | DEBUG | Full JWK (public keys are meant to be public) |
| Policy evaluation outcome | INFO | `{rule_name, outcome, sub, act}` |
| Client credentials | **NEVER** | — |
| Request body | DEBUG | Truncated to first 256 bytes, or hash only |
| Error details | ERROR | Error code + message, never stack traces in production |

### R-4: AI Output Rules

The AI must never output:
- Private key material in any conversation or code
- Real secrets from `.env` files (even if the human pastes them)
- Full Oath Tokens in examples (use truncated format: `eyJhbGci...`)
- Test secrets that look like real secrets (no `sk_live_*` patterns)

For examples, use these obviously-fake values:
- Keys: `EXAMPLE_KEY_DO_NOT_USE`
- Tokens: `eyJhbGciOiJFZERTQSIsInR5cCI6Im9tK2p3dCJ9.EXAMPLE.SIGNATURE`
- URLs: `https://issuer.example.oathmesh.dev`

### R-5: Environment Tier Handling

| Environment | Secret Storage | Access |
|---|---|---|
| Local dev | `.env` file (not committed) | Developer only |
| CI/CD | GitHub Actions secrets / CI env vars | Pipeline only |
| Staging | Cloud secret manager (SSM, Secret Manager, Key Vault) | Deployment service only |
| Production | Cloud KMS/HSM for signing keys, Cloud secret manager for config | Deployment service only, audited access |

### R-6: Dependency Vulnerability Policy

When a dependency vulnerability is discovered:

1. **Critical/High CVE**: Create a hotfix task immediately (`workflows/hotfix.md`). Update the dependency within 48 hours.
2. **Medium CVE**: Add to current sprint. Update within 2 weeks.
3. **Low CVE**: Add to backlog. Update in next dependency maintenance window.

### R-7: Secret Rotation

If a secret is suspected of being exposed:

1. **Immediately**: Rotate the secret (new key, new credential)
2. **Within 1 hour**: Audit access logs for the exposed secret
3. **Within 24 hours**: Determine root cause and prevent recurrence
4. **Document**: Write an incident report and update relevant security policies
