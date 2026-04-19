"""Strawberry GraphQL Middleware for OathMesh

Pre-execution authentication and rate limiting for GraphQL operations.
Integrates with OathMesh SDK for token verification and field authorization.

Usage with Strawberry and FastAPI:

```python
from strawberry import Schema
from strawberry.fastapi import GraphQLRouter
from oathmesh.middleware.graphql import create_oathmesh_middleware
from oathmesh.middleware.types import OathMeshGraphQLConfig

config = OathMeshGraphQLConfig(
    audience="https://api.example.com",
    trusted_issuers=["https://issuer.oathmesh.tech"],
)

router = GraphQLRouter(
    schema,
    middleware=[create_oathmesh_middleware(config)],
)
```
"""

from __future__ import annotations
import time
import asyncio
from typing import Optional, Callable, Any
from functools import wraps
import logging

from oathmesh.verify import verify_token, VerifierConfig
from oathmesh.errors import OathMeshError
from .types import OathMeshGraphQLContext, OathMeshGraphQLConfig, RateLimitBucket

logger = logging.getLogger(__name__)

# In-memory rate limiting store
_rate_limit_store: dict[str, RateLimitBucket] = {}
_rate_limit_lock = asyncio.Lock()

DEFAULT_LIMITS = {
    'queries_per_minute': 100,
    'mutations_per_minute': 10,
}


async def check_rate_limit(
    subject: str,
    operation_type: str,
    limits: OathMeshGraphQLConfig,
) -> Optional[str]:
    """Check if the request exceeds rate limits.
    
    Args:
        subject: Token subject for rate limit key
        operation_type: 'query' or 'mutation'
        limits: Configuration with rate limit settings
        
    Returns:
        Error code if rate limit exceeded, None if allowed.
    """
    async with _rate_limit_lock:
        now = time.time()
        limit = (
            limits.queries_per_minute
            if operation_type == 'query'
            else limits.mutations_per_minute
        )

        bucket = _rate_limit_store.get(subject)

        # Create new window if needed
        if bucket is None or (now - bucket.window_start) >= 60:
            bucket = RateLimitBucket()
            _rate_limit_store[subject] = bucket

        current_count = bucket.request_counts.get(operation_type, 0)

        if current_count >= limit:
            return 'rate_limit_exceeded'

        bucket.request_counts[operation_type] = current_count + 1
        return None


def get_operation_type(operation: Optional[Any]) -> str:
    """Determine the operation type (query or mutation).
    
    Args:
        operation: GraphQL operation object
        
    Returns:
        'mutation' or 'query' (default)
    """
    if operation and hasattr(operation, 'operation'):
        if operation.operation == 'mutation':
            return 'mutation'
    return 'query'


def extract_oath_token(auth_header: Optional[str]) -> Optional[str]:
    """Extract OathMesh token from Authorization header.
    
    Accepts "OathMesh <token>" format.
    
    Args:
        auth_header: Authorization header value
        
    Returns:
        Token string or None
    """
    if not auth_header:
        return None
    if auth_header.startswith('OathMesh '):
        token = auth_header[9:]
        if token.strip():  # Ensure token is not just whitespace
            return token
    return None


def create_oathmesh_middleware(config: OathMeshGraphQLConfig):
    """Create Strawberry GraphQL middleware for OathMesh authentication.
    
    This middleware:
    1. Extracts JWT from Authorization header
    2. Verifies token using OathMesh SDK
    3. Applies rate limiting per operation type
    4. Injects verified claims into GraphQL context
    
    Args:
        config: OathMesh GraphQL middleware configuration
        
    Returns:
        Strawberry middleware callable
        
    Example:
        ```python
        from strawberry.fastapi import GraphQLRouter
        from oathmesh.middleware.graphql import create_oathmesh_middleware
        
        middleware = create_oathmesh_middleware(config)
        router = GraphQLRouter(
            schema,
            middleware=[middleware],
        )
        ```
    """
    
    async def middleware(app: Any, handler: Callable, info: Any) -> Any:
        """Middleware handler for request context.
        
        Args:
            app: Strawberry app instance
            handler: Next handler in middleware chain
            info: GraphQL resolve info and context
            
        Returns:
            Result from handler
            
        Raises:
            OathMeshError: On authentication or rate limit failures
        """
        # Get request from context if available
        request = getattr(info.context, 'request', None) if hasattr(info, 'context') else None
        
        if not request:
            # No request object — skip authentication
            return await handler(info)

        # Extract authorization header
        auth_header = None
        if hasattr(request, 'headers'):
            if isinstance(request.headers, dict):
                auth_header = request.headers.get('authorization')
            else:
                auth_header = request.headers.get('authorization')

        token = extract_oath_token(auth_header)

        if not token:
            raise OathMeshError(
                'claim_missing:token',
                'missing or invalid Authorization header',
                'provide a token in the format "Authorization: OathMesh <token>"',
            )

        # Create verifier config
        verifier_config = VerifierConfig(
            audience=config.audience,
            trusted_issuers=config.trusted_issuers,
            require_request_binding=config.require_request_binding,
            replay_cache=config.replay_cache,
            policy_evaluator=config.policy_evaluator,
        )

        # Verify token
        try:
            claims = verify_token(f'OathMesh {token}', verifier_config)
        except OathMeshError:
            raise
        except Exception as err:
            raise OathMeshError(
                'verification_failed',
                str(err),
                'check token and issuer configuration',
            )

        # Get operation type
        operation = getattr(info, 'operation', None)
        operation_type = get_operation_type(operation)

        # Check rate limit
        subject = claims.principal.subject
        rate_limit_error = await check_rate_limit(subject, operation_type, config)

        if rate_limit_error:
            if config.on_rate_limit_exceeded:
                config.on_rate_limit_exceeded(subject, operation_type)
            raise OathMeshError(
                'rate_limit_exceeded',
                f'rate limit exceeded for {operation_type} operations',
                'wait before retrying',
            )

        # Inject OathMesh context
        oathmesh_context = OathMeshGraphQLContext(
            claims=claims,
            verified=True,
            rate_limit={
                'remaining': 0,
                'reset_at': time.time() + 60,
            },
        )

        # Add to context
        if not hasattr(info.context, 'oathmesh'):
            info.context.oathmesh = oathmesh_context

        # Call next handler
        return await handler(info)

    return middleware


def get_oathmesh_context(context: Any) -> Optional[OathMeshGraphQLContext]:
    """Get OathMesh context from a GraphQL resolver.
    
    Use this in field resolvers to access verified claims.
    
    Args:
        context: GraphQL context object
        
    Returns:
        OathMeshGraphQLContext if available, None otherwise
        
    Example:
        ```python
        from oathmesh.middleware.graphql import get_oathmesh_context
        
        @strawberry.field
        async def current_user(self, info: Info) -> User:
            oathmesh = get_oathmesh_context(info.context)
            if not oathmesh or not oathmesh.verified:
                raise Exception("Not authenticated")
            return await db.get_user(oathmesh.claims.principal.subject)
        ```
    """
    if context is None:
        return None

    if isinstance(context, OathMeshGraphQLContext):
        return context

    return getattr(context, 'oathmesh', None)
