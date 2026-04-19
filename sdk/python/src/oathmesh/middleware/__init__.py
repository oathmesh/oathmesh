"""OathMesh GraphQL Middleware for Python

Complete GraphQL middleware support for Python/Strawberry with:
- Pre-execution JWT verification
- Rate limiting (queries/mutations)
- Field-level authorization decorators
- Context injection for resolver access

Example usage:

```python
import strawberry
from strawberry.asgi import make_graphql_router
from oathmesh.middleware.graphql import create_oathmesh_middleware
from oathmesh.middleware.types import OathMeshGraphQLConfig
from oathmesh.middleware.decorators import oathmesh_requires

# Configure middleware
config = OathMeshGraphQLConfig(
    audience="https://api.example.com",
    trusted_issuers=["https://issuer.oathmesh.tech"],
    queries_per_minute=100,
    mutations_per_minute=10,
)

# Define schema with protected fields
@strawberry.type
class User:
    id: strawberry.ID
    name: str
    
    @oathmesh_requires(require="action:read:user:email")
    async def email(self) -> str:
        return self._email

@strawberry.type
class Query:
    @strawberry.field
    async def current_user(self, info: Info) -> User:
        from oathmesh.middleware.graphql import get_oathmesh_context
        oathmesh = get_oathmesh_context(info.context)
        if not oathmesh:
            raise Exception("Not authenticated")
        return await get_user(oathmesh.claims.principal.subject)

schema = strawberry.Schema(query=Query)

# Create router with middleware
from fastapi import FastAPI
app = FastAPI()
middleware = create_oathmesh_middleware(config)
graphql_router = make_graphql_router(
    schema,
    middleware=[middleware],
)
app.include_router(graphql_router, prefix="/graphql")
```
"""

from .graphql import (
    create_oathmesh_middleware,
    get_oathmesh_context,
    extract_oath_token,
    check_rate_limit,
)
from .decorators import (
    oathmesh_requires,
    requires_authenticated,
    check_field_permission,
    PermissionDenied,
)
from .types import (
    OathMeshGraphQLContext,
    OathMeshGraphQLConfig,
    RateLimitBucket,
)

__all__ = [
    # Middleware
    'create_oathmesh_middleware',
    'get_oathmesh_context',
    'extract_oath_token',
    'check_rate_limit',
    # Decorators
    'oathmesh_requires',
    'requires_authenticated',
    'check_field_permission',
    'PermissionDenied',
    # Types
    'OathMeshGraphQLContext',
    'OathMeshGraphQLConfig',
    'RateLimitBucket',
]
