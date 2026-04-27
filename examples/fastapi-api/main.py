"""
examples/fastapi-api — FastAPI example with OathMesh verification.

Usage:
    cd examples/fastapi-api
    pip install -r requirements.txt
    OATHMESH_AUDIENCE=https://inventory.internal \
    OATHMESH_TRUSTED_ISSUERS=https://issuer.oathmesh.tech \
    uvicorn main:app --host 0.0.0.0 --port 8000
"""

import os
import logging

from fastapi import FastAPI, Request, Depends, HTTPException
from oathmesh import verify_token, VerifierConfig, OathMeshError, InMemoryRevocationCache
from oathmesh.types import VerifiedCallerContext

logger = logging.getLogger("oathmesh.example")
logging.basicConfig(level=logging.INFO)

app = FastAPI(
    title="OathMesh FastAPI Example",
    description="Inventory API protected by OathMesh token verification",
)

audience = os.environ.get("OATHMESH_AUDIENCE", "https://inventory.internal")
issuers = os.environ.get("OATHMESH_TRUSTED_ISSUERS", "http://localhost:4000").split(",")

revocation_cache = InMemoryRevocationCache()
revocation_cache.revoke("agent://test/revoked-svc")

config = VerifierConfig(
    audience=audience,
    trusted_issuers=[i.strip() for i in issuers],
    revocation_list=revocation_cache,
    on_denied=lambda err: logger.warning("denied: %s — %s", err.code, err.message),
    on_verified=lambda ctx: logger.info("allowed: %s → %s", ctx.principal.subject, ctx.action),
)


async def require_oathmesh(request: Request) -> VerifiedCallerContext:
    """FastAPI dependency that verifies the OathMesh token."""
    auth = request.headers.get("authorization", "")
    try:
        return verify_token(auth, config)
    except OathMeshError as e:
        raise HTTPException(status_code=401, detail=e.to_dict())


@app.get("/healthz")
async def healthz():
    return {"status": "ok"}


@app.get("/inventory")
async def get_inventory(caller: VerifiedCallerContext = Depends(require_oathmesh)):
    return {
        "status": "success",
        "data": ["widget-a", "widget-b", "widget-c"],
        "caller": {
            "subject": caller.principal.subject,
            "action": caller.action,
            "token_id": caller.token_id,
            "environment": caller.environment,
        },
    }


@app.post("/inventory")
async def create_item(caller: VerifiedCallerContext = Depends(require_oathmesh)):
    return {
        "status": "created",
        "created_by": caller.principal.subject,
    }
