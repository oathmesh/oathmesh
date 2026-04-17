"""OathMesh — Machine-call identity verification SDK for Python.

Usage with FastAPI::

    from fastapi import FastAPI, Request, Depends, HTTPException
    from oathmesh import verify_token, VerifierConfig, OathMeshError

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
        return {"subject": caller.principal.subject}
"""

from .errors import OathMeshError
from .types import VerifiedCallerContext, Principal, Source
from .verify import VerifierConfig, verify_token, verify_raw_token, extract_token
from .client import OathMeshClient

__all__ = [
    "OathMeshError",
    "VerifiedCallerContext",
    "VerifierConfig",
    "Principal",
    "Source",
    "verify_token",
    "verify_raw_token",
    "extract_token",
    "OathMeshClient",
]

__version__ = "1.0.0"
