# Quickstart: Protect a FastAPI Service

**Time:** ~5 minutes

## Prerequisites

- Python 3.9+
- A running OathMesh issuer

## Step 1: Install

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
    on_denied=lambda err: print(f"denied: {err.code}"),
)

async def require_oathmesh(request: Request):
    try:
        return verify_token(request.headers.get("authorization", ""), config)
    except OathMeshError as e:
        raise HTTPException(status_code=401, detail=e.to_dict())
```

## Step 3: Protect Your Endpoints

```python
@app.get("/inventory")
async def get_inventory(caller=Depends(require_oathmesh)):
    return {
        "subject": caller.principal.subject,
        "action": caller.action,
        "token_id": caller.token_id,
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
```

## Next Steps

- [GitHub Actions to internal API](github-actions-to-internal-api.md)
- [Run the full demo](local-demo-docker-compose.md)
