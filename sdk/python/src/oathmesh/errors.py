"""OathMesh error types matching the error taxonomy."""

from __future__ import annotations
from typing import Optional


# All valid error codes from the OathMesh error taxonomy
ErrorCode = str  # One of the values below:
# claim_missing:token, claim_missing:iss, claim_missing:sub,
# claim_missing:aud, claim_missing:act, claim_missing:jti,
# signature_invalid, issuer_untrusted, token_expired,
# audience_mismatch, algorithm_not_allowed, replay_detected,
# policy_denied, binding_mismatch, binding_required, subject_revoked, verification_failed


class OathMeshError(Exception):
    """Structured error returned on verification failure.

    Attributes:
        code: Machine-readable error code from the taxonomy.
        message: Human-readable description of what went wrong.
        fix: Actionable instruction for how to resolve the error.
    """

    def __init__(self, code: str, message: str, fix: Optional[str] = None):
        self.code = code
        self.message = message
        self.fix = fix
        super().__init__(f"{code}: {message}")

    def to_dict(self) -> dict:
        """Serialize to the standard OathMesh error JSON shape."""
        result: dict = {"error": self.code, "message": self.message}
        if self.fix:
            result["fix"] = self.fix
        return result
