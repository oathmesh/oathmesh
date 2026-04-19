← [Back to Index](../INDEX.md)

# Tutorial: Protect a FastAPI Service

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

⏱️ **Time**: ~5 minutes | 📋 **Prerequisites**: Python 3.9+, running OathMesh issuer | 🎯 **Outcome**: FastAPI/Flask service with OathMesh token verification

---

> 🆕 **New here?** Start with [Getting Started](../GETTING_STARTED.md) for a guided introduction.

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
    trusted_issuers=["https://issuer.oathmesh.tech"],
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

## Troubleshooting

| Issue | Likely Cause | Fix |
|---|---|---|
| `401 {"detail":{"code":"issuer_untrusted"}}` | Trusted issuer list mismatch | Update `trusted_issuers` to exact issuer URL |
| `binding_mismatch` | `rqh` present but request hash differs | Either remove `rqh` for this flow or compute canonical request hash correctly |
| Import/runtime errors for SDK | Wrong environment or package not installed | Use Python 3.9+ and reinstall with `pip install oathmesh` |

## Flask

```python
from flask import Flask, request, jsonify
from oathmesh import verify_token, VerifierConfig, OathMeshError

app = Flask(__name__)
config = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.oathmesh.tech"],
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
- [Protect an Express API](protect-express-api.md)
- [Protect a Next.js API](protect-nextjs-api.md)
- [Protect a Go chi API](protect-chi-api.md)

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [Python SDK](../../sdk/python/README.md) | Full SDK reference |
| [Verification Rules](../protocol/verification-rules.md) | 14-step pipeline details |
| [Error Taxonomy](../protocol/error-taxonomy.md) | All error codes |
| [Threat Model](../security/threat-model.md) | Security model |
