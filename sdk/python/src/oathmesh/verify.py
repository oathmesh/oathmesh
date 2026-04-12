import jwt
from jwt import PyJWKClient
from typing import List, Optional, Dict, Any
from .errors import OathMeshError

class VerifierConfig:
    def __init__(self, audience: str, trusted_issuers: List[str]):
        self.audience = audience
        self.trusted_issuers = trusted_issuers

def verify_token(token: str, config: VerifierConfig) -> Dict[str, Any]:
    if not token or not token.startswith("OathMesh "):
        raise OathMeshError("claim_missing:token", "missing or invalid Authorization header")
    
    token = token[len("OathMesh "):]
    
    try:
        unverified_header = jwt.get_unverified_header(token)
        unverified_claims = jwt.decode(token, options={"verify_signature": False})
    except jwt.DecodeError:
        raise OathMeshError("verification_failed", "malformed token")

    iss = unverified_claims.get("iss")
    if not iss or iss not in config.trusted_issuers:
        raise OathMeshError("unknown_issuer", "issuer not trusted")

    # Fetch and configure JWKS
    jwks_url = f"{iss}/.well-known/jwks.json"
    jwks_client = PyJWKClient(jwks_url, cache_keys=True)
    
    try:
        signing_key = jwks_client.get_signing_key_from_jwt(token)
    except jwt.exceptions.PyJWKClientError:
        raise OathMeshError("verification_failed", "failed to resolve JWKS key")

    try:
        data = jwt.decode(
            token,
            signing_key.key,
            algorithms=["EdDSA"],
            audience=config.audience,
            leeway=10
        )
    except jwt.ExpiredSignatureError:
        raise OathMeshError("token_expired", "token expired")
    except jwt.InvalidAudienceError:
        raise OathMeshError("audience_mismatch", "wrong audience")
    except Exception as e:
        raise OathMeshError("verification_failed", str(e))

    if not data.get("act"):
        raise OathMeshError("claim_missing:act", "missing act")
    if not data.get("sub"):
        raise OathMeshError("claim_missing:sub", "missing sub")
    if not data.get("jti"):
        raise OathMeshError("claim_missing:jti", "missing jti")

    return {
        "principal": {
            "subject": data["sub"],
            "issuer": data["iss"]
        },
        "action": data["act"],
        "token_id": data["jti"],
        "environment": data.get("env", "")
    }
