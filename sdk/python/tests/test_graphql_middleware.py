from __future__ import annotations

from types import SimpleNamespace
from unittest.mock import Mock

import pytest

from oathmesh.errors import OathMeshError
from oathmesh.middleware import graphql as graphql_module
from oathmesh.middleware.decorators import oathmesh_requires, requires_authenticated
from oathmesh.middleware.graphql import create_oathmesh_middleware
from oathmesh.middleware.types import OathMeshGraphQLConfig
from oathmesh.types import Principal, VerifiedCallerContext


def _claims(subject: str, scope: list[str] | None = None) -> VerifiedCallerContext:
    return VerifiedCallerContext(
        principal=Principal(issuer="https://issuer.test.local", subject=subject),
        action="read",
        token_id=f"jti-{subject}",
        environment="test",
        scope=scope or ["action:read:user"],
    )


def _config(**kwargs) -> OathMeshGraphQLConfig:
    return OathMeshGraphQLConfig(
        audience="https://api.test.local",
        trusted_issuers=["https://issuer.test.local"],
        **kwargs,
    )


@pytest.mark.asyncio
async def test_graphql_middleware_auth_header_handling():
    middleware = create_oathmesh_middleware(_config())

    async def next_handler(_info):
        return "ok"

    missing = SimpleNamespace(
        context=SimpleNamespace(request=SimpleNamespace(headers={})),
        operation=SimpleNamespace(operation="query"),
    )
    with pytest.raises(OathMeshError) as missing_err:
        await middleware(None, next_handler, missing)
    assert missing_err.value.code == "claim_missing:token"

    wrong_scheme = SimpleNamespace(
        context=SimpleNamespace(request=SimpleNamespace(headers={"authorization": "Bearer abc"})),
        operation=SimpleNamespace(operation="query"),
    )
    with pytest.raises(OathMeshError) as scheme_err:
        await middleware(None, next_handler, wrong_scheme)
    assert scheme_err.value.code == "claim_missing:token"


@pytest.mark.asyncio
async def test_graphql_middleware_injects_context_from_verified_token(monkeypatch: pytest.MonkeyPatch):
    captured = {}

    def fake_verify_token(auth_header, verifier_config):
        captured["auth_header"] = auth_header
        captured["audience"] = verifier_config.audience
        return _claims("svc://graphql/python")

    monkeypatch.setattr(graphql_module, "verify_token", fake_verify_token)
    middleware = create_oathmesh_middleware(_config())

    async def next_handler(info):
        return info.context.oathmesh.claims.principal.subject

    info = SimpleNamespace(
        context=SimpleNamespace(request=SimpleNamespace(headers={"authorization": "OathMesh token-123"})),
        operation=SimpleNamespace(operation="query"),
    )

    result = await middleware(None, next_handler, info)
    assert result == "svc://graphql/python"
    assert captured["auth_header"] == "OathMesh token-123"
    assert captured["audience"] == "https://api.test.local"
    assert info.context.oathmesh.verified is True


@pytest.mark.asyncio
async def test_graphql_middleware_rate_limit_behavior(monkeypatch: pytest.MonkeyPatch):
    graphql_module._rate_limit_store.clear()

    monkeypatch.setattr(graphql_module, "verify_token", lambda *_: _claims("svc://graphql/rate-limit"))
    on_rate_limit_exceeded = Mock()
    middleware = create_oathmesh_middleware(
        _config(queries_per_minute=1, on_rate_limit_exceeded=on_rate_limit_exceeded)
    )

    async def next_handler(_info):
        return "ok"

    first = SimpleNamespace(
        context=SimpleNamespace(request=SimpleNamespace(headers={"authorization": "OathMesh token"})),
        operation=SimpleNamespace(operation="query"),
    )
    second = SimpleNamespace(
        context=SimpleNamespace(request=SimpleNamespace(headers={"authorization": "OathMesh token"})),
        operation=SimpleNamespace(operation="query"),
    )

    assert await middleware(None, next_handler, first) == "ok"
    with pytest.raises(OathMeshError) as err:
        await middleware(None, next_handler, second)
    assert err.value.code == "rate_limit_exceeded"
    on_rate_limit_exceeded.assert_called_once_with("svc://graphql/rate-limit", "query")


@pytest.mark.asyncio
async def test_graphql_middleware_auth_and_permission_integration(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr(graphql_module, "verify_token", lambda *_: _claims("svc://graphql/integration"))
    middleware = create_oathmesh_middleware(_config())

    @oathmesh_requires(require="action:read:user")
    def resolver(info):
        return "allowed"

    async def next_handler(info):
        return resolver(info)

    info = SimpleNamespace(
        context=SimpleNamespace(request=SimpleNamespace(headers={"authorization": "OathMesh token"})),
        operation=SimpleNamespace(operation="query"),
    )

    assert await middleware(None, next_handler, info) == "allowed"


def test_graphql_permission_decorators_enforce_permissions():
    @oathmesh_requires(require="action:admin")
    def admin_only(info):
        return "secret"

    denied_info = SimpleNamespace(
        context=SimpleNamespace(oathmesh=SimpleNamespace(claims=_claims("svc://graphql/denied"), verified=True))
    )
    assert admin_only(denied_info) is None

    @requires_authenticated()
    def authenticated_only(info):
        return "allowed"

    unauthenticated = SimpleNamespace(context=SimpleNamespace(oathmesh=SimpleNamespace(verified=False)))
    authenticated = SimpleNamespace(
        context=SimpleNamespace(oathmesh=SimpleNamespace(claims=_claims("svc://graphql/auth"), verified=True))
    )

    assert authenticated_only(unauthenticated) is None
    assert authenticated_only(authenticated) == "allowed"
