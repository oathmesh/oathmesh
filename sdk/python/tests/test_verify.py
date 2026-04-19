"""Tests for OathMesh Python SDK."""

import pytest
from oathmesh import verify_token, extract_token, VerifierConfig, OathMeshError
from oathmesh.types import VerifiedCallerContext, Principal


# ─── extract_token ────────────────────────────────────────────────────────────

class TestExtractToken:
    def test_oathmesh_prefix(self):
        assert extract_token("OathMesh abc.def.ghi") == "abc.def.ghi"

    def test_none(self):
        assert extract_token(None) is None

    def test_empty(self):
        assert extract_token("") is None

    def test_unknown_scheme(self):
        assert extract_token("Basic abc123") is None
        assert extract_token("Token abc123") is None


# ─── OathMeshError ────────────────────────────────────────────────────────────

class TestOathMeshError:
    def test_attributes(self):
        err = OathMeshError("audience_mismatch", "wrong audience", "fix it")
        assert err.code == "audience_mismatch"
        assert err.message == "wrong audience"
        assert err.fix == "fix it"
        assert isinstance(err, Exception)

    def test_to_dict(self):
        err = OathMeshError("token_expired", "expired", "mint new")
        assert err.to_dict() == {
            "error": "token_expired",
            "message": "expired",
            "fix": "mint new",
        }

    def test_to_dict_no_fix(self):
        err = OathMeshError("token_expired", "expired")
        d = err.to_dict()
        assert "fix" not in d

    def test_str_includes_code(self):
        err = OathMeshError("replay_detected", "already used")
        assert "replay_detected" in str(err)


# ─── verify_token ─────────────────────────────────────────────────────────────

class TestVerifyToken:
    CONFIG = VerifierConfig(
        audience="https://inventory.internal",
        trusted_issuers=["http://issuer.local"],
    )

    def test_missing_header(self):
        with pytest.raises(OathMeshError) as exc_info:
            verify_token("", self.CONFIG)
        assert exc_info.value.code == "claim_missing:token"
        assert exc_info.value.fix is not None

    def test_wrong_scheme(self):
        with pytest.raises(OathMeshError) as exc_info:
            verify_token("Basic abc123", self.CONFIG)
        assert exc_info.value.code == "claim_missing:token"

    def test_malformed_token(self):
        with pytest.raises(OathMeshError) as exc_info:
            verify_token("OathMesh not-a-real-token", self.CONFIG)
        assert exc_info.value.code in ("verification_failed", "algorithm_not_allowed", "claim_missing")

    def test_on_denied_hook_fires(self):
        errors = []
        config = VerifierConfig(
            audience="https://inventory.internal",
            trusted_issuers=["http://issuer.local"],
            on_denied=lambda e: errors.append(e),
        )
        with pytest.raises(OathMeshError):
            verify_token("", config)
        assert len(errors) == 1
        assert errors[0].code == "claim_missing:token"


# ─── Types ────────────────────────────────────────────────────────────────────

class TestTypes:
    def test_principal_frozen(self):
        p = Principal(issuer="https://iss.dev", subject="svc://a/b")
        assert p.issuer == "https://iss.dev"
        with pytest.raises(AttributeError):
            p.issuer = "changed"  # type: ignore

    def test_verified_context_defaults(self):
        ctx = VerifiedCallerContext(
            principal=Principal(issuer="i", subject="s"),
            action="read",
            token_id="jti-123",
        )
        assert ctx.environment == ""
        assert ctx.scope == []
        assert ctx.reason is None
        assert ctx.source is None
