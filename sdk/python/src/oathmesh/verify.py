"""OathMesh token verification — framework-agnostic core.

Verifies OathMesh tokens following the 14-step verification pipeline.
Supports Ed25519 (EdDSA) and ES256 algorithms.
"""

from __future__ import annotations

import jwt
from jwt import PyJWKClient
from typing import List, Optional, Callable, Any
from .errors import OathMeshError
from .types import VerifiedCallerContext, Principal, Source

# Module-level JWKS client cache — shared across calls within the same process.
# In serverless (Lambda, Cloud Functions), module scope persists across warm invocations.
_jwks_clients: dict[str, PyJWKClient] = {}

ALLOWED_ALGORITHMS = ["EdDSA", "ES256"]


class VerifierConfig:
    """Configuration for the OathMesh verifier.

    Args:
        audience: The audience URL this receiver expects (exact match).
        trusted_issuers: Trusted issuer URLs (explicit allowlist).
        on_verified: Optional callback on every successful verification.
        on_denied: Optional callback on every denied request.
    """

    def __init__(
        self,
        audience: str,
        trusted_issuers: List[str],
        on_verified: Optional[Callable[[VerifiedCallerContext], None]] = None,
        on_denied: Optional[Callable[[OathMeshError], None]] = None,
    ):
        self.audience = audience
        self.trusted_issuers = trusted_issuers
        self.on_verified = on_verified
        self.on_denied = on_denied


def extract_token(auth_header: Optional[str]) -> Optional[str]:
    """Extract the raw token string from an Authorization header.

    Accepts:
      - ``OathMesh <token>`` (canonical)
      - ``Bearer <token>`` (compatibility mode)

    Returns None if the header is missing or uses an unknown scheme.
    """
    if not auth_header:
        return None
    if auth_header.startswith("OathMesh "):
        return auth_header[9:]
    if auth_header.startswith("Bearer "):
        return auth_header[7:]
    return None


def verify_token(auth_header: str, config: VerifierConfig) -> VerifiedCallerContext:
    """Verify an OathMesh token from the Authorization header value.

    Args:
        auth_header: The full Authorization header (e.g., "OathMesh <token>").
        config: Verifier configuration.

    Returns:
        VerifiedCallerContext on success.

    Raises:
        OathMeshError: On any verification failure, with a machine-readable
        ``code``, human-readable ``message``, and actionable ``fix``.
    """
    raw_token = extract_token(auth_header)
    if not raw_token:
        err = OathMeshError(
            "claim_missing:token",
            "missing or invalid Authorization header",
            "provide a token in the format 'Authorization: OathMesh <token>'",
        )
        if config.on_denied:
            config.on_denied(err)
        raise err

    return verify_raw_token(raw_token, config)


def verify_raw_token(token: str, config: VerifierConfig) -> VerifiedCallerContext:
    """Verify a raw OathMesh token string (without the scheme prefix).

    Use this when you've already extracted the token from the header.
    """
    # Step 02: Decode and validate header
    try:
        unverified_header = jwt.get_unverified_header(token)
    except jwt.DecodeError:
        _deny(config, "verification_failed", "malformed token header", "check token format")

    typ = unverified_header.get("typ", "")
    if typ != "om+jwt":
        _deny(config, "algorithm_not_allowed", f'invalid token type "{typ}"', "token typ must be om+jwt")

    alg = unverified_header.get("alg", "")
    if alg == "none":
        _deny(config, "algorithm_not_allowed", 'algorithm "none" is rejected', "use EdDSA or ES256")
    if alg not in ALLOWED_ALGORITHMS:
        _deny(config, "algorithm_not_allowed", f'algorithm "{alg}" is not allowed', "use EdDSA or ES256")

    # Step 03-04: Extract issuer and check trust
    try:
        unverified_claims = jwt.decode(token, options={"verify_signature": False})
    except jwt.DecodeError:
        _deny(config, "verification_failed", "malformed token payload", "check token format")

    iss = unverified_claims.get("iss")
    if not iss:
        _deny(config, "claim_missing:iss", "missing iss claim", "include iss when minting")
    if iss not in config.trusted_issuers:
        _deny(config, "issuer_untrusted", f'issuer "{iss}" is not trusted', "add it to trusted_issuers")

    # Step 05-06: Load JWKS and verify signature
    jwks_client = _get_jwks_client(iss)

    try:
        signing_key = jwks_client.get_signing_key_from_jwt(token)
    except Exception:
        _deny(config, "signature_invalid", "failed to resolve JWKS signing key", "check issuer JWKS availability")

    try:
        data = jwt.decode(
            token,
            signing_key.key,
            algorithms=ALLOWED_ALGORITHMS,
            audience=config.audience,
            leeway=10,
        )
    except jwt.ExpiredSignatureError:
        _deny(config, "token_expired", "token has expired", "mint a new token")
    except jwt.InvalidAudienceError:
        _deny(config, "audience_mismatch", "token audience does not match", f"mint with aud={config.audience}")
    except jwt.InvalidIssuerError:
        _deny(config, "issuer_untrusted", "issuer claim verification failed", "check trusted_issuers")
    except jwt.InvalidSignatureError:
        _deny(config, "signature_invalid", "JWS signature verification failed", "check issuer key")
    except Exception as e:
        _deny(config, "verification_failed", str(e), "check token format")

    # Step 11: Verify required claims
    if not data.get("sub"):
        _deny(config, "claim_missing:sub", "missing sub claim", "include sub when minting")
    if not data.get("act"):
        _deny(config, "claim_missing:act", "missing act claim", "include act when minting")
    if not data.get("jti"):
        _deny(config, "claim_missing:jti", "missing jti claim", "jti is auto-generated by the issuer")

    # Build source provenance if present
    source = None
    src_data = data.get("src")
    if isinstance(src_data, dict) and src_data.get("type"):
        source = Source(
            type=src_data["type"],
            repo=src_data.get("repo"),
            workflow=src_data.get("workflow"),
            run_id=src_data.get("run_id"),
            sha=src_data.get("sha"),
        )

    ctx = VerifiedCallerContext(
        principal=Principal(issuer=iss, subject=data["sub"]),
        action=data["act"],
        token_id=data["jti"],
        environment=data.get("env", ""),
        scope=data.get("scope", []),
        reason=data.get("reason"),
        source=source,
    )

    if config.on_verified:
        config.on_verified(ctx)

    return ctx


def _get_jwks_client(issuer: str) -> PyJWKClient:
    """Get or create a cached JWKS client for the given issuer."""
    if issuer not in _jwks_clients:
        jwks_url = f"{issuer}/.well-known/jwks.json"
        _jwks_clients[issuer] = PyJWKClient(jwks_url, cache_keys=True)
    return _jwks_clients[issuer]


def _deny(config: VerifierConfig, code: str, message: str, fix: str) -> None:
    """Raise an OathMeshError and call the on_denied hook."""
    err = OathMeshError(code, message, fix)
    if config.on_denied:
        config.on_denied(err)
    raise err
