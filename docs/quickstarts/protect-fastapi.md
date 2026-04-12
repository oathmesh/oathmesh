# Quickstart: Protect a FastAPI Service

**Time:** ~5 minutes

This guide adds OathMesh token verification to a FastAPI service using the `oathmesh` Python SDK.

## Prerequisites

- Python 3.8+
- A running OathMesh issuer

## Step 1: Install the SDK

```bash
pip install oathmesh
```

## Step 2: Create a Verification Dependency

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
```

## Step 3: Protect Your Endpoints

```python
@app.get("/inventory")
async def get_inventory(caller=Depends(require_oathmesh)):
    return {
        "caller": caller["principal"]["subject"],
        "action": caller["action"],
    }
```

## Step 4: Test It

```bash
TOKEN=$(oathmesh mint \
  --sub "job://ci/nightly-sync" \
  --aud "https://inventory.internal" \
  --act "inventory.read" \
  --quiet)

curl -H "Authorization: OathMesh $TOKEN" http://localhost:8000/inventory
```

Expected output:
```json
{"caller":"job://ci/nightly-sync","action":"inventory.read"}
```

## Next Steps

- [GitHub Actions to internal API](github-actions-to-internal-api.md)
- [Run the full demo](local-demo-docker-compose.md)
