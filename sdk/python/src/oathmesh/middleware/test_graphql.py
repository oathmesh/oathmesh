"""Tests for OathMesh GraphQL Middleware

Comprehensive test suite covering authentication, rate limiting, and field authorization.
"""

import pytest
import asyncio
import time
from unittest.mock import Mock, AsyncMock, MagicMock
from typing import Optional, Any

from oathmesh.types import VerifiedCallerContext, Principal
from oathmesh.errors import OathMeshError
from oathmesh.middleware.graphql import (
    create_oathmesh_middleware,
    get_oathmesh_context,
    extract_oath_token,
    check_rate_limit,
    get_operation_type,
)
from oathmesh.middleware.decorators import (
    oathmesh_requires,
    requires_authenticated,
    check_field_permission,
    PermissionDenied,
)
from oathmesh.middleware.types import (
    OathMeshGraphQLContext,
    OathMeshGraphQLConfig,
)


# ─── Test Utilities ──────────────────────────────────────────────────────────

def create_mock_context(
    subject: str = "svc://test",
    scope: Optional[list[str]] = None,
    verified: bool = True,
) -> OathMeshGraphQLContext:
    """Create a mock OathMesh context for testing."""
    if scope is None:
        scope = ["action:read"]
    
    claims = VerifiedCallerContext(
        principal=Principal(
            issuer="https://issuer.test.local",
            subject=subject,
        ),
        action="read",
        token_id="jti-test-123",
        environment="test",
        scope=scope,
    )
    
    return OathMeshGraphQLContext(
        claims=claims,
        verified=verified,
    )


# ─── Tests ───────────────────────────────────────────────────────────────────

class TestGetOathMeshContext:
    """Test get_oathmesh_context helper function."""
    
    def test_returns_none_when_context_is_none(self):
        """Test that None context returns None."""
        assert get_oathmesh_context(None) is None
    
    def test_returns_none_when_no_oathmesh_property(self):
        """Test that context without oathmesh property returns None."""
        context = Mock()
        context.oathmesh = None
        result = get_oathmesh_context(context)
        assert result is None
    
    def test_returns_context_when_present(self):
        """Test that oathmesh context is returned when present."""
        mock_context = create_mock_context()
        context = Mock()
        context.oathmesh = mock_context
        
        result = get_oathmesh_context(context)
        assert result is not None
        assert result.verified is True
        assert result.claims.principal.subject == "svc://test"


class TestExtractOathToken:
    """Test token extraction from Authorization header."""
    
    def test_extracts_oath_mesh_prefix(self):
        """Test extraction of OathMesh prefixed token."""
        header = "OathMesh abc.def.ghi"
        token = extract_oath_token(header)
        assert token == "abc.def.ghi"
    
    def test_returns_none_for_missing_header(self):
        """Test that None is returned for missing header."""
        assert extract_oath_token(None) is None
        assert extract_oath_token("") is None
    
    def test_returns_none_for_unknown_scheme(self):
        """Test that None is returned for unknown auth scheme."""
        assert extract_oath_token("Bearer token123") is None
        assert extract_oath_token("Basic abc123") is None
    
    def test_returns_none_for_malformed_header(self):
        """Test that None is returned for malformed header."""
        assert extract_oath_token("OathMesh") is None  # Missing token
        assert extract_oath_token("OathMesh ") is None  # Only whitespace


class TestCheckFieldPermission:
    """Test field permission checking."""
    
    def test_returns_false_for_none_context(self):
        """Test that None context always denies."""
        assert check_field_permission(None, "action:read") is False
    
    def test_returns_false_when_not_verified(self):
        """Test that unverified context denies access."""
        oathmesh_context = create_mock_context(verified=False)
        context = Mock()
        context.oathmesh = oathmesh_context
        assert check_field_permission(context, "action:read") is False
    
    def test_returns_true_for_exact_scope_match(self):
        """Test exact scope matching."""
        oathmesh_context = create_mock_context(scope=["action:read:email", "action:write:profile"])
        context = Mock()
        context.oathmesh = oathmesh_context
        assert check_field_permission(context, "action:read:email") is True
    
    def test_returns_false_for_missing_scope(self):
        """Test that missing scope is denied."""
        oathmesh_context = create_mock_context(scope=["action:read"])
        context = Mock()
        context.oathmesh = oathmesh_context
        assert check_field_permission(context, "action:write") is False
    
    def test_matches_wildcard_requirements(self):
        """Test wildcard scope matching."""
        oathmesh_context = create_mock_context(scope=["action:read:email", "action:read:profile"])
        context = Mock()
        context.oathmesh = oathmesh_context
        assert check_field_permission(context, "action:*") is True
    
    def test_handles_empty_scopes(self):
        """Test that empty scope list denies access."""
        oathmesh_context = create_mock_context(scope=[])
        context = Mock()
        context.oathmesh = oathmesh_context
        assert check_field_permission(context, "action:read") is False


class TestOathMeshRequiresDecorator:
    """Test @oathmesh_requires field authorization decorator."""
    
    def test_denies_access_without_permission(self):
        """Test that resolver returns None when permission is denied."""
        @oathmesh_requires(require="action:admin")
        def resolver(info):
            return "secret"
        
        # Mock info object with context wrapping OathMeshGraphQLContext
        oathmesh_context = create_mock_context(scope=["action:read"])
        info = Mock()
        info.context = Mock()
        info.context.oathmesh = oathmesh_context
        
        result = resolver(info)
        assert result is None
    
    def test_grants_access_with_permission(self):
        """Test that resolver executes when permission is granted."""
        @oathmesh_requires(require="action:read")
        def resolver(info):
            return "secret"
        
        # Mock info object with context wrapping OathMeshGraphQLContext
        oathmesh_context = create_mock_context(scope=["action:read"])
        info = Mock()
        info.context = Mock()
        info.context.oathmesh = oathmesh_context
        
        result = resolver(info)
        assert result == "secret"
    
    @pytest.mark.asyncio
    async def test_async_resolver_with_permission(self):
        """Test async resolver with permission granted."""
        @oathmesh_requires(require="action:read")
        async def resolver(info):
            return "async-secret"
        
        oathmesh_context = create_mock_context(scope=["action:read"])
        info = Mock()
        info.context = Mock()
        info.context.oathmesh = oathmesh_context
        
        result = await resolver(info)
        assert result == "async-secret"
    
    @pytest.mark.asyncio
    async def test_async_resolver_without_permission(self):
        """Test async resolver returns None when denied."""
        @oathmesh_requires(require="action:admin")
        async def resolver(info):
            return "admin-secret"
        
        oathmesh_context = create_mock_context(scope=["action:read"])
        info = Mock()
        info.context = Mock()
        info.context.oathmesh = oathmesh_context
        
        result = await resolver(info)
        assert result is None
        info.context = create_mock_context(scope=["action:read"])
        
        result = await resolver(info)
        assert result == "async-secret"
    
    @pytest.mark.asyncio
    async def test_async_resolver_without_permission(self):
        """Test async resolver returns None when denied."""
        @oathmesh_requires(require="action:admin")
        async def resolver(info):
            return "admin-secret"
        
        info = Mock()
        info.context = create_mock_context(scope=["action:read"])
        
        result = await resolver(info)
        assert result is None


class TestRequiresAuthenticatedDecorator:
    """Test @requires_authenticated decorator."""
    
    def test_denies_unverified_user(self):
        """Test that unverified users get None."""
        @requires_authenticated()
        def resolver(info):
            return "protected"
        
        info = Mock()
        info.context = create_mock_context(verified=False)
        
        result = resolver(info)
        assert result is None
    
    def test_grants_verified_user(self):
        """Test that verified users get access."""
        @requires_authenticated()
        def resolver(info):
            return "protected"
        
        info = Mock()
        info.context = create_mock_context(verified=True)
        
        result = resolver(info)
        assert result == "protected"
    
    @pytest.mark.asyncio
    async def test_async_authenticated_access(self):
        """Test async resolver with authentication."""
        @requires_authenticated()
        async def resolver(info):
            return "async-protected"
        
        info = Mock()
        info.context = create_mock_context(verified=True)
        
        result = await resolver(info)
        assert result == "async-protected"


class TestRateLimiting:
    """Test rate limiting functionality."""
    
    @pytest.mark.asyncio
    async def test_allows_requests_within_limit(self):
        """Test that requests within limit are allowed."""
        config = OathMeshGraphQLConfig(
            audience="https://api.test.local",
            trusted_issuers=["https://issuer.test.local"],
            queries_per_minute=2,
        )
        
        # Clear rate limit store
        from oathmesh.middleware import graphql as gql_module
        gql_module._rate_limit_store.clear()
        
        subject = "svc://test"
        
        # First request should be allowed
        error = await check_rate_limit(subject, "query", config)
        assert error is None
        
        # Second request should be allowed
        error = await check_rate_limit(subject, "query", config)
        assert error is None
    
    @pytest.mark.asyncio
    async def test_denies_requests_over_limit(self):
        """Test that requests over limit are denied."""
        config = OathMeshGraphQLConfig(
            audience="https://api.test.local",
            trusted_issuers=["https://issuer.test.local"],
            queries_per_minute=1,
        )
        
        from oathmesh.middleware import graphql as gql_module
        gql_module._rate_limit_store.clear()
        
        subject = "svc://test"
        
        # First request
        error = await check_rate_limit(subject, "query", config)
        assert error is None
        
        # Second request should exceed limit
        error = await check_rate_limit(subject, "query", config)
        assert error == 'rate_limit_exceeded'
    
    @pytest.mark.asyncio
    async def test_separate_limits_for_queries_and_mutations(self):
        """Test that queries and mutations have separate limits."""
        config = OathMeshGraphQLConfig(
            audience="https://api.test.local",
            trusted_issuers=["https://issuer.test.local"],
            queries_per_minute=1,
            mutations_per_minute=1,
        )
        
        from oathmesh.middleware import graphql as gql_module
        gql_module._rate_limit_store.clear()
        
        subject = "svc://test"
        
        # Query up to limit
        assert await check_rate_limit(subject, "query", config) is None
        assert await check_rate_limit(subject, "query", config) == 'rate_limit_exceeded'
        
        # Mutation should work (separate limit)
        assert await check_rate_limit(subject, "mutation", config) is None


class TestGetOperationType:
    """Test operation type detection."""
    
    def test_detects_query_operation(self):
        """Test detection of query operation."""
        operation = Mock()
        operation.operation = "query"
        assert get_operation_type(operation) == "query"
    
    def test_detects_mutation_operation(self):
        """Test detection of mutation operation."""
        operation = Mock()
        operation.operation = "mutation"
        assert get_operation_type(operation) == "mutation"
    
    def test_defaults_to_query_for_none(self):
        """Test that None operation defaults to query."""
        assert get_operation_type(None) == "query"
    
    def test_defaults_to_query_for_missing_operation(self):
        """Test that missing operation attribute defaults to query."""
        operation = Mock(spec=[])  # No attributes
        assert get_operation_type(operation) == "query"


class TestMiddlewareIntegration:
    """Integration tests for middleware."""
    
    def test_context_sharing_between_resolvers(self):
        """Test that context is shared between resolvers."""
        shared_context = create_mock_context()
        context_obj = Mock()
        context_obj.oathmesh = shared_context
        
        # First resolver accesses context
        first = get_oathmesh_context(context_obj)
        assert first is not None
        
        # Second resolver accesses same context
        second = get_oathmesh_context(context_obj)
        assert second is not None
        
        # Both have same reference
        assert first is second
    
    def test_nested_field_authorization(self):
        """Test authorization across nested fields."""
        context = create_mock_context(scope=["action:read:user", "action:read:user:email"])
        
        # Parent field: user (requires action:read:user)
        parent_allowed = check_field_permission(context, "action:read:user")
        assert parent_allowed is True
        
        # Child field: email (requires action:read:user:email)
        child_allowed = check_field_permission(context, "action:read:user:email")
        assert child_allowed is True
    
    def test_wildcard_scope_cascade(self):
        """Test wildcard scope matching at different levels."""
        context = create_mock_context(scope=["action:*"])
        
        assert check_field_permission(context, "action:read") is True
        assert check_field_permission(context, "action:write") is True
        assert check_field_permission(context, "action:delete") is True
        assert check_field_permission(context, "other:read") is False


class TestErrorHandling:
    """Test error handling in middleware."""
    
    def test_permission_denied_exception(self):
        """Test PermissionDenied exception."""
        with pytest.raises(PermissionDenied):
            raise PermissionDenied("Access denied")
    
    def test_permission_denied_message(self):
        """Test custom error message."""
        try:
            raise PermissionDenied("Custom message")
        except PermissionDenied as e:
            assert str(e.message) == "Custom message"


class TestConfigurationDefaults:
    """Test default configuration values."""
    
    def test_default_rate_limits(self):
        """Test that defaults are applied."""
        config = OathMeshGraphQLConfig(
            audience="https://api.test.local",
            trusted_issuers=["https://issuer.test.local"],
        )
        
        assert config.queries_per_minute == 100
        assert config.mutations_per_minute == 10
    
    def test_custom_rate_limits(self):
        """Test custom rate limit configuration."""
        config = OathMeshGraphQLConfig(
            audience="https://api.test.local",
            trusted_issuers=["https://issuer.test.local"],
            queries_per_minute=50,
            mutations_per_minute=5,
        )
        
        assert config.queries_per_minute == 50
        assert config.mutations_per_minute == 5


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
