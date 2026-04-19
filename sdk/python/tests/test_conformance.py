import jwt
import pytest
import time

from oathmesh import OathMeshError
from oathmesh.types import InMemoryReplayCache
from oathmesh.verify import VerifierConfig, verify_raw_token, verify_token
import oathmesh.verify as verify_module


CONFIG = VerifierConfig(
    audience="https://inventory.internal",
    trusted_issuers=["https://issuer.local"],
)


def _base_claims(**overrides):
    now = int(time.time())
    claims = {
        "iss": "https://issuer.local",
        "sub": "svc://python/conformance",
        "aud": "https://inventory.internal",
        "act": "read",
        "iat": now,
        "exp": now + 120,
        "jti": "jti-python-conformance-1",
    }
    claims.update(overrides)
    return claims


def _stub_signing_key():
    class Key:
        key = "fake-key"
    return Key()


def test_token_parsing_validation_failures(monkeypatch: pytest.MonkeyPatch):
    def bad_header(_token):
        raise jwt.DecodeError("bad header")

    monkeypatch.setattr(verify_module.jwt, "get_unverified_header", bad_header)

    with pytest.raises(OathMeshError) as exc:
        verify_raw_token("not-a-token", CONFIG)
    assert exc.value.code == "claim_missing"


def test_issuer_check_untrusted(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr(verify_module.jwt, "get_unverified_header", lambda _t: {"typ": "om+jwt", "alg": "EdDSA"})
    monkeypatch.setattr(
        verify_module.jwt,
        "decode",
        lambda _t, options=None, **kwargs: _base_claims(iss="https://evil.local") if options else {},
    )

    with pytest.raises(OathMeshError) as exc:
        verify_raw_token("token.part.sig", CONFIG)
    assert exc.value.code == "issuer_untrusted"


def test_audience_check_mismatch(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr(verify_module.jwt, "get_unverified_header", lambda _t: {"typ": "om+jwt", "alg": "EdDSA"})

    def fake_decode(_token, key=None, algorithms=None, audience=None, leeway=None, issuer=None, options=None):
        if options and options.get("verify_signature") is False:
            return _base_claims()
        raise jwt.InvalidAudienceError("audience mismatch")

    monkeypatch.setattr(verify_module.jwt, "decode", fake_decode)
    monkeypatch.setattr(
        verify_module,
        "_get_jwks_client",
        lambda _issuer: type("FakeClient", (), {"get_signing_key_from_jwt": lambda self, _token: _stub_signing_key()})(),
    )

    with pytest.raises(OathMeshError) as exc:
        verify_raw_token("token.part.sig", CONFIG)
    assert exc.value.code == "audience_mismatch"


def test_replay_detection_semantics(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr(verify_module.jwt, "get_unverified_header", lambda _t: {"typ": "om+jwt", "alg": "EdDSA"})

    def fake_decode(_token, key=None, algorithms=None, audience=None, leeway=None, issuer=None, options=None):
        if options and options.get("verify_signature") is False:
            return _base_claims()
        return _base_claims()

    monkeypatch.setattr(verify_module.jwt, "decode", fake_decode)
    monkeypatch.setattr(
        verify_module,
        "_get_jwks_client",
        lambda _issuer: type("FakeClient", (), {"get_signing_key_from_jwt": lambda self, _token: _stub_signing_key()})(),
    )

    cfg = VerifierConfig(
        audience="https://inventory.internal",
        trusted_issuers=["https://issuer.local"],
        replay_cache=InMemoryReplayCache(default_ttl=300),
    )

    verify_raw_token("token.part.sig", cfg)
    with pytest.raises(OathMeshError) as exc:
        verify_raw_token("token.part.sig", cfg)
    assert exc.value.code == "replay_detected"


def test_middleware_auth_header_handling_semantics():
    with pytest.raises(OathMeshError) as missing:
        verify_token("", CONFIG)
    assert missing.value.code == "claim_missing:token"

    with pytest.raises(OathMeshError) as wrong_scheme:
        verify_token("Bearer abc123", CONFIG)
    assert wrong_scheme.value.code == "claim_missing:token"


def test_alg_none_rejection(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr(verify_module.jwt, "get_unverified_header", lambda _t: {"typ": "om+jwt", "alg": "none"})
    with pytest.raises(OathMeshError) as exc:
        verify_raw_token("token.part.sig", CONFIG)
    assert exc.value.code == "algorithm_not_allowed"


def test_subject_format_validation(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr(verify_module.jwt, "get_unverified_header", lambda _t: {"typ": "om+jwt", "alg": "EdDSA"})
    monkeypatch.setattr(
        verify_module.jwt,
        "decode",
        lambda _t, options=None, **kwargs: _base_claims(sub="invalid-subject") if options else {},
    )
    with pytest.raises(OathMeshError) as exc:
        verify_raw_token("token.part.sig", CONFIG)
    assert exc.value.code == "claim_missing:sub"


def test_binding_required_semantics(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr(verify_module.jwt, "get_unverified_header", lambda _t: {"typ": "om+jwt", "alg": "EdDSA"})

    def fake_decode(_token, key=None, algorithms=None, audience=None, leeway=None, issuer=None, options=None):
        if options and options.get("verify_signature") is False:
            return _base_claims()
        return _base_claims()

    monkeypatch.setattr(verify_module.jwt, "decode", fake_decode)
    monkeypatch.setattr(
        verify_module,
        "_get_jwks_client",
        lambda _issuer: type("FakeClient", (), {"get_signing_key_from_jwt": lambda self, _token: _stub_signing_key()})(),
    )

    cfg = VerifierConfig(
        audience="https://inventory.internal",
        trusted_issuers=["https://issuer.local"],
        require_request_binding=True,
    )
    with pytest.raises(OathMeshError) as exc:
        verify_raw_token("token.part.sig", cfg)
    assert exc.value.code == "binding_required"


def test_iat_future_rejection(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr(verify_module.jwt, "get_unverified_header", lambda _t: {"typ": "om+jwt", "alg": "EdDSA"})

    def fake_decode(_token, key=None, algorithms=None, audience=None, leeway=None, issuer=None, options=None):
        if options and options.get("verify_signature") is False:
            return _base_claims(iat=int(time.time()) + 60)
        return _base_claims(iat=int(time.time()) + 60)

    monkeypatch.setattr(verify_module.jwt, "decode", fake_decode)
    monkeypatch.setattr(
        verify_module,
        "_get_jwks_client",
        lambda _issuer: type("FakeClient", (), {"get_signing_key_from_jwt": lambda self, _token: _stub_signing_key()})(),
    )

    with pytest.raises(OathMeshError) as exc:
        verify_raw_token("token.part.sig", CONFIG)
    assert exc.value.code == "token_expired"
