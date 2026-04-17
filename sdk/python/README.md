# oathmesh

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  <b>OathMesh token verification SDK for Python</b> — fully typed, framework-agnostic.
</p>

<p align="center">
  <a href="https://pypi.org/project/oathmesh/">
    <img src="https://img.shields.io/pypi/v/oathmesh.svg" alt="PyPI version">
  </a>
  <a href="https://pypi.org/project/oathmesh/">
    <img src="https://img.shields.io/pypi/dm/oathmesh" alt="PyPI downloads">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml">
    <img src="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/oathmesh/oathmesh" alt="License">
  </a>
</p>

---

## Installation

```bash
pip install oathmesh
# or
poetry add oathmesh
# or
pip install -e .[test]  # Development
```

### Requirements

- Python 3.9+
- PyJWT >= 2.8.0
- cryptography >= 41.0.0

---

## Quick Start

```python
from oathmesh import verify_token, VerifierConfig, OathMeshError

config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.dev"],
)

# In your FastAPI/Flask route
caller = verify_token(request.headers.get("authorization", ""), config)
print(caller.principal.subject)  # "agent://repo/acme/deploy-bot"
```

---

## Framework Support

| Framework | Integration Method | Example |
|-----------|-------------------|---------|
| **FastAPI** | Dependency | `Depends(require_oathmesh)` |
| **Flask** | Before request | `@app.before_request` |
| **Django** | Decorator | `@require_oathmesh` |
| **Any** | Core verifier | `verify_raw_token()` |

---

## FastAPI

```python
from fastapi import FastAPI, Request, Depends, HTTPException
from oathmesh import verify_token, VerifierConfig, OathMeshError

app = FastAPI()

config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.dev"],
)

async def require_oathmesh(request: Request):
    try:
        return verify_token(request.headers.get("authorization", ""), config)
    except OathMeshError as e:
        raise HTTPException(status_code=401, detail=e.to_dict())

@app.get("/inventory")
async def get_inventory(caller=Depends(require_oathmesh)):
    return {
        "subject": caller.principal.subject,
        "action": caller.action,
    }
```

## Flask

```python
from flask import Flask, request, jsonify
from oathmesh import verify_token, VerifierConfig, OathMeshError

app = Flask(__name__)

config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.dev"],
)

@app.before_request
def check_oathmesh():
    try:
        request.oathmesh = verify_token(
            request.headers.get("Authorization", ""), config
        )
    except OathMeshError as e:
        return jsonify(e.to_dict()), 401

@app.get("/inventory")
def get_inventory():
    return {"subject": request.oathmesh.principal.subject}
```

## Django (View Decorator)

```python
from functools import wraps
from django.http import JsonResponse
from oathmesh import verify_token, VerifierConfig, OathMeshError

config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.dev"],
)

def require_oathmesh(view_func):
    @wraps(view_func)
    def wrapper(request, *args, **kwargs):
        try:
            request.oathmesh = verify_token(
                request.META.get("HTTP_AUTHORIZATION", ""), config
            )
        except OathMeshError as e:
            return JsonResponse(e.to_dict(), status=401)
        return view_func(request, *args, **kwargs)
    return wrapper
```

## OathMeshClient (Auto-Minting)

If you need to construct signed tokens to make upstream requests as a service, use the `OathMeshClient`:

```python
from oathmesh import OathMeshClient
import urllib.request

client = OathMeshClient(
    issuer="https://issuer.oathmesh.dev",
    api_key="your_api_key"
)

# Auto-fetches and caches tokens efficiently until TTL
token = client.mint(
    sub="service://inventory/worker",
    aud="https://financial.internal",
    act="payment.process"
)

req = urllib.request.Request("https://financial.internal/pay")
req.add_header("Authorization", f"OathMesh {token}")
response = urllib.request.urlopen(req)
```

---

## Core Verifier (Framework-agnostic)

```python
from oathmesh import verify_raw_token, extract_token, VerifierConfig

config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.dev"],
)

# Extract token from any header format
token = extract_token(auth_header)       # extracting exactly from "OathMesh ..."

# Verify
caller = verify_raw_token(token, config)
print(caller.principal.subject)          # "agent://repo/acme/deploy-bot"
print(caller.action)                     # "inventory.write"
print(caller.token_id)                   # UUID
```

---

## Lifecycle Hooks

```python
config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.dev"],
    on_verified=lambda ctx: logger.info("allowed", extra={"sub": ctx.principal.subject}),
    on_denied=lambda err: logger.warning("denied", extra={"code": err.code}),
)
```

---

## API Reference

### `verify_token(auth_header, config) → VerifiedCallerContext`

Verify from a full Authorization header value.

### `verify_raw_token(token, config) → VerifiedCallerContext`

Verify from a raw token string (no prefix).

### `extract_token(auth_header) → str | None`

Extract token from exactly `OathMesh ` prefixed headers.

### `VerifiedCallerContext`

Frozen dataclass:

```python
@dataclass(frozen=True)
class VerifiedCallerContext:
    principal: Principal        # .issuer, .subject
    action: str
    token_id: str
    environment: str = ""
    scope: list[str] = []
    reason: str | None = None
    source: Source | None = None
```

### `OathMeshError`

```python
class OathMeshError(Exception):
    code: str       # e.g., "audience_mismatch"
    message: str    # human-readable
    fix: str | None # actionable instruction
    step: int | None # execution step number (e.g. 11)

    def to_dict(self) -> dict:
        ...
```

---

## Development

```bash
pip install -e .[test]
pytest tests/
```
