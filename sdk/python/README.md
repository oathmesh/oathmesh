# oathmesh

OathMesh verification SDK for Python. Verify Oath Tokens in FastAPI, Flask, or any Python HTTP framework.

## Installation

```bash
pip install oathmesh
```

## Usage with FastAPI

```python
from fastapi import FastAPI, Request, Depends, HTTPException
from oathmesh import verify_token, VerifierConfig, OathMeshError

app = FastAPI()

config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.dev"],
)

async def require_oathmesh(request: Request):
    auth = request.headers.get("authorization", "")
    try:
        return verify_token(auth, config)
    except OathMeshError as e:
        raise HTTPException(status_code=401, detail={"error": e.code, "message": e.message})

@app.get("/inventory")
async def get_inventory(caller=Depends(require_oathmesh)):
    return {"caller": caller["principal"]["subject"]}
```

## API

### `verify_token(authorization_header, config)`

Verifies an OathMesh token from the Authorization header.

**Parameters:**
- `authorization_header` (str) — The full `Authorization` header value (e.g., `"OathMesh <token>"`).
- `config` (VerifierConfig) — Configuration with audience and trusted issuers.

**Returns:** A dict with verified caller context:
```python
{
    "principal": {"subject": "agent://repo/acme/bot", "issuer": "https://..."},
    "action": "deploy",
    "token_id": "550e8400-...",
    "environment": "prod"
}
```

**Raises:** `OathMeshError` with `.code` and `.message` on verification failure.

### `VerifierConfig(audience, trusted_issuers)`

Configuration for the verifier.

- `audience` (str) — The audience URL this receiver expects.
- `trusted_issuers` (list[str]) — Trusted issuer URLs.

### `OathMeshError`

Exception with `.code` (machine-readable) and `.message` (human-readable).

See the [Error Taxonomy](../../docs/protocol/error-taxonomy.md) for all error codes.
