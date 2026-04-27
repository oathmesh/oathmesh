"""OathMesh token verification — framework-agnostic core.

Verifies OathMesh tokens following the 14-step verification pipeline.
Supports Ed25519 (EdDSA) and ES256 algorithms.
"""

from __future__ import annotations

import hashlib
import re
import time
import jwt
from jwt import PyJWKClient
from typing import List, Optional, Callable
from .errors import OathMeshError
from .types import VerifiedCallerContext, Principal, Source, ReplayCache, PolicyEvaluator, RevocationList

# Module-level JWKS client cache — shared across calls within the same process.
# In serverless (Lambda, Cloud Functions), module scope persists across warm invocations.
_jwks_clients: dict[str, PyJWKClient] = {}

ALLOWED_ALGORITHMS = ["EdDSA", "ES256"]
SUBJECT_RE = re.compile(r"^(agent|svc|job|tool|user)://[a-zA-Z0-9/_.-]{1,256}$")
DEFAULT_CLOCK_SKEW_LEEWAY = 30
MAX_EXP_UNIX = 4102444800


class VerifierConfig:
    """Configuration for the OathMesh verifier.

    Args:
        audience: The audience URL this receiver expects (exact match).
        trusted_issuers: Trusted issuer URLs (explicit allowlist).
        require_request_binding: Enforce that tokens MUST include an rqh claim.
            When True, tokens without rqh are rejected with error "binding_required".
            Recommended for all write/mutate endpoints to prevent tampering.
            Default: False (for backward compatibility).
        replay_cache: Optional replay cache for preventing token reuse.
            Use InMemoryReplayCache for development, or implement Redis-based cache for production.
            Default: None (no replay checking).
        policy_evaluator: Optional policy evaluator for authorization decisions.
            Use JsonPolicyEvaluator with a JSON policy document.
            Default: None (no policy enforcement).
        on_verified: Optional callback on every successful verification.
        on_denied: Optional callback on every denied request.
    """

    def __init__(
        self,
        audience: str,
        trusted_issuers: List[str],
        require_request_binding: bool = False,
        request_hash: Optional[str] = None,
        clock_skew_leeway: int = DEFAULT_CLOCK_SKEW_LEEWAY,
        replay_cache: Optional[ReplayCache] = None,
        revocation_list: Optional[RevocationList] = None,
        policy_evaluator: Optional[PolicyEvaluator] = None,
        on_verified: Optional[Callable[[VerifiedCallerContext], None]] = None,
        on_denied: Optional[Callable[[OathMeshError], None]] = None,
    ):
        self.audience = audience
        self.trusted_issuers = trusted_issuers
        self.require_request_binding = require_request_binding
        self.request_hash = request_hash
        self.clock_skew_leeway = clock_skew_leeway
        self.replay_cache = replay_cache
        self.revocation_list = revocation_list
        self.policy_evaluator = policy_evaluator
        self.on_verified = on_verified
        self.on_denied = on_denied


def extract_token(auth_header: Optional[str]) -> Optional[str]:
    """Extract the raw token string from an Authorization header.

    Accepts:
      - ``OathMesh <token>`` (canonical)

    Returns None if the header is missing or uses an unknown scheme.
    """
    if not auth_header:
        return None
    if auth_header.startswith("OathMesh "):
        return auth_header[9:]
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
    parts = token.split(".")
    if len(parts) != 3:
        _deny(
            config,
            "claim_missing",
            f"invalid token format: expected 3 segments, got {len(parts)}",
            "provide a valid OathMesh token in compact JWS format (header.payload.signature)",
        )

    # Step 02: Decode and validate header
    try:
        unverified_header = jwt.get_unverified_header(token)
    except jwt.DecodeError:
        _deny(config, "claim_missing", "failed to decode token header", "provide a valid base64url-encoded token header")

    typ = unverified_header.get("typ", "")
    if typ != "om+jwt":
        _deny(config, "claim_missing", f'token type "{typ}" is not valid — expected "om+jwt"', "token typ must be om+jwt")

    alg = unverified_header.get("alg", "")
    if alg == "none":
        _deny(config, "algorithm_not_allowed", 'algorithm "none" is rejected', "use EdDSA or ES256")
    if alg not in ALLOWED_ALGORITHMS:
        _deny(config, "algorithm_not_allowed", f'algorithm "{alg}" is not allowed', "use EdDSA or ES256")

    # Step 03-04: Extract issuer and check trust
    try:
        unverified_claims = jwt.decode(token, options={"verify_signature": False})
    except jwt.DecodeError:
        _deny(config, "claim_missing", "failed to decode token payload", "provide a valid base64url-encoded token payload")

    iss = unverified_claims.get("iss")
    if not iss:
        _deny(config, "claim_missing:iss", "missing iss claim", "include iss when minting")
    if iss not in config.trusted_issuers:
        _deny(config, "issuer_untrusted", f'issuer "{iss}" is not trusted', "add it to trusted_issuers")

    # Step 11 (moved early): verify required claims.
    _required_string_claim(config, unverified_claims, "sub")
    _required_string_claim(config, unverified_claims, "aud")
    _required_string_claim(config, unverified_claims, "act")
    _required_number_claim(config, unverified_claims, "iat")
    _required_number_claim(config, unverified_claims, "exp")
    _required_string_claim(config, unverified_claims, "jti")

    # Step 11.5: subject format
    if not SUBJECT_RE.match(str(unverified_claims["sub"])):
        _deny(
            config,
            "claim_missing:sub",
            f'invalid subject format: "{unverified_claims["sub"]}"',
            "subject must match schema (svc://, agent://, job://, tool://, user:// plus allowed chars)"
        )

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
            issuer=iss,
            leeway=config.clock_skew_leeway,
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

    # Step 07: post-signature issuer trust re-check.
    if data.get("iss") not in config.trusted_issuers:
        _deny(config, "issuer_untrusted", f'issuer "{data.get("iss")}" failed post-signature trust check', "verify trusted_issuers configuration")

    # Ensure signed payload still carries required claims.
    if not data.get("sub"):
        _deny(config, "claim_missing:sub", "missing sub claim", "include sub when minting")
    if not data.get("act"):
        _deny(config, "claim_missing:act", "missing act claim", "include act when minting")
    if not data.get("jti"):
        _deny(config, "claim_missing:jti", "missing jti claim", "jti is auto-generated by the issuer")

    # Step 08-10: temporal and audience checks aligned with Go.
    now = int(time.time())
    exp = int(data["exp"])
    if exp > MAX_EXP_UNIX:
        _deny(config, "token_expired", "token expiry is invalid or too far in the future", "mint a token with sane exp")
    if now > exp + config.clock_skew_leeway:
        _deny(config, "token_expired", "token has expired", "mint a new token")

    iat = int(data["iat"])
    if iat > now + config.clock_skew_leeway:
        _deny(config, "token_expired", "token issued-at is in the future", "check clock synchronization between issuer and receiver")

    if data.get("nbf") is not None and int(data["nbf"]) > now + config.clock_skew_leeway:
        _deny(config, "token_expired", "token not-before is in the future", "token cannot be used yet")

    if data.get("aud") != config.audience:
        _deny(config, "audience_mismatch", "token audience does not match", f"mint with aud={config.audience}")

    # Step 12: verify request hash binding when both are present.
    if data.get("rqh") and config.request_hash:
        expected_hash = "sha256:" + hashlib.sha256(config.request_hash.encode("utf-8")).hexdigest()
        if data["rqh"] != expected_hash:
            _deny(
                config,
                "binding_mismatch",
                f'request hash mismatch: token has "{data["rqh"]}" but request hash is "{expected_hash}"',
                "ensure the request body has not been modified since token minting",
            )

    # Step 12b: Enforce rqh if RequireRequestBinding is set
    if config.require_request_binding and not data.get("rqh"):
        _deny(
            config,
            "binding_required",
            "token missing rqh (request hash) claim",
            "mint a token with rqh= sha256:<canonical-request> for write/mutate operations"
        )

    # Step 13: Check replay cache (if configured)
    if config.replay_cache:
        jti = data.get("jti")
        if jti and config.replay_cache.check(jti):
            _deny(
                config,
                "replay_detected",
                f"token {jti} has already been used",
                "each Oath Token can only be used once — mint a new token"
            )
        if jti:
            config.replay_cache.add(jti)

    # Step 13.5: optional revocation list.
    if config.revocation_list and config.revocation_list.is_revoked(str(data["sub"])):
        _deny(config, "subject_revoked", f'subject {data["sub"]} has been revoked', "mint a token for an active subject")

    # Step 14: Evaluate policy (if configured)
    if config.policy_evaluator:
        src_data = data.get("src")
        src_type = src_repo = src_wflow = None
        if isinstance(src_data, dict):
            src_type = src_data.get("type")
            src_repo = src_data.get("repo")
            src_wflow = src_data.get("workflow")
        policy_input = PolicyInput(
            iss=iss,
            sub=data["sub"],
            aud=data["aud"],
            act=data["act"],
            scope=data.get("scope"),
            env=data.get("env"),
            tenant=data.get("tenant"),
            src_type=src_type,
            src_repo=src_repo,
            src_wflow=src_wflow,
        )
        decision = config.policy_evaluator.evaluate(policy_input)
        if decision.outcome == "deny":
            _deny(
                config,
                "policy_denied",
                decision.deny_reason or "policy evaluation denied",
                "check policy rules for this request"
            )

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
        tenant=data.get("tenant", ""),
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


def _required_string_claim(config: VerifierConfig, claims: dict, name: str) -> str:
    value = claims.get(name)
    if not isinstance(value, str) or value == "":
        _deny(config, f"claim_missing:{name}", f"missing {name} claim", f"include {name} when minting")
    return value


def _required_number_claim(config: VerifierConfig, claims: dict, name: str) -> int:
    value = claims.get(name)
    if not isinstance(value, (int, float)):
        _deny(config, f"claim_missing:{name}", f"missing {name} claim", f"include {name} when minting")
    return int(value)
