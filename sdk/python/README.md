---
title: Python SDK
description: OathMesh token verification SDK for Python with FastAPI, Flask, and Django integration patterns.
keywords: [oathmesh, python, fastapi, flask, django, jwt, machine-identity]
audience: application-developers, platform-engineers, enterprise-teams
---

[Docs Index](../../docs/INDEX.md)

# oathmesh

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

OathMesh token verification SDK for Python.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Framework Integrations](#framework-integrations)
- [Configuration](#configuration)
- [Error Handling](#error-handling)
- [Troubleshooting](#troubleshooting)
- [Security Notes](#security-notes)
- [Production Tips](#production-tips)
- [Related Docs](#related-docs)

## Installation

```bash
pip install oathmesh
# or
poetry add oathmesh
```

Requirements:
- Python 3.9+
- PyJWT >= 2.8.0
- cryptography >= 41.0.0

## Quick Start

```python
from oathmesh import verify_token, VerifierConfig

config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.tech"],
)

caller = verify_token("OathMesh <token>", config)
print(caller.principal.subject)
```

Canonical header:

```http
Authorization: OathMesh <token>
```

## Framework Integrations

### FastAPI

Use a dependency that calls `verify_token(...)` and raises `HTTPException(401)` on `OathMeshError`.

### Flask

Use `@app.before_request` to verify `request.headers.get("Authorization", "")`.

### Django

Use a view decorator that verifies `request.META.get("HTTP_AUTHORIZATION", "")`.

## Configuration

```python
from oathmesh import VerifierConfig, verify_raw_token, extract_token

config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.tech"],
    on_verified=lambda ctx: logger.info("allowed", extra={"sub": ctx.principal.subject}),
    on_denied=lambda err: logger.warning("denied", extra={"code": err.code, "step": err.step}),
)

token = extract_token(auth_header)
caller = verify_raw_token(token, config)
```

Verifier semantics:
- Canonical and required header for verification is `Authorization: OathMesh <token>`.
- `extract_token` returns `None` for non-OathMesh schemes.
- If upstream sends `Bearer`, translate it to `OathMesh` before calling verifier APIs.
- Verifier behavior is aligned with canonical Go step semantics for conformance-critical checks (including `alg=none` rejection, subject format validation, binding-required semantics, and future-`iat` rejection).
- Parity is behavioral across language runtimes; implementations are intentionally language-native, not byte-identical.
- `revocation_list` remains optional. Cross-SDK conformance currently marks Python revocation behavior as SKIP/N/A.

## Error Handling

```python
from oathmesh import OathMeshError

try:
    # verify...
    pass
except OathMeshError as err:
    print(err.code, err.message, err.fix, err.step)
```

Common codes: `claim_missing:token`, `issuer_untrusted`, `audience_mismatch`, `signature_invalid`, `token_expired`.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `claim_missing:token` | Missing/invalid header | Send `Authorization: OathMesh <token>` |
| `issuer_untrusted` | `iss` not in `trusted_issuers` | Add exact issuer URL |
| `audience_mismatch` | Token `aud` differs from `audience` | Mint token with matching `aud` |
| `token_expired` | TTL elapsed / clock skew | Mint fresh token and sync clocks |
| `signature_invalid` | JWKS/issuer mismatch | Check issuer URL and JWKS reachability |

## Security Notes

- Verify only tokens from trusted issuers you control.
- Never log raw tokens or API keys.
- Keep issuer TLS and JWKS endpoints reachable and authenticated.
- Treat all claims as untrusted until verification succeeds.

## Production Tips

- Reuse verifier config objects across requests.
- Log `error code` and `step` for operational triage.
- Use short TTLs and rotate issuer keys safely.
- Monitor deny rates by error code.

## Related Docs

- [Getting Started](../../docs/GETTING_STARTED.md)
- [Troubleshooting Guide](../../docs/TROUBLESHOOTING.md)
- [Community](../../docs/COMMUNITY.md)
- [Enterprise Guide](../../docs/enterprise/README.md)
