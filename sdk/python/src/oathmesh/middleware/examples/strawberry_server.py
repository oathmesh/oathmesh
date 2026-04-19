"""Strawberry GraphQL Server Example with OathMesh Middleware

This example demonstrates:
1. Setting up Strawberry GraphQL with FastAPI
2. Registering OathMesh middleware for authentication
3. Using decorators for field-level authorization
4. Accessing verified claims in resolvers

Usage:
    # Install dependencies (if not already installed)
    pip install fastapi uvicorn strawberry-graphql[asgi] oathmesh

    # Run the server
    python examples/strawberry_server.py

Testing:
    # Query without token → 401 error
    curl -X POST http://localhost:8000/graphql \\
      -H "Content-Type: application/json" \\
      -d '{"query": "{ currentUser { id name } }"}'

    # Query with token
    curl -X POST http://localhost:8000/graphql \\
      -H "Content-Type: application/json" \\
      -H "Authorization: OathMesh <token>" \\
      -d '{"query": "{ currentUser { id name email } }"}'
"""

import strawberry
from typing import Optional
from datetime import datetime
from fastapi import FastAPI, Request
from strawberry.asgi import GraphQL
from strawberry.types import Info

from oathmesh.middleware.graphql import (
    create_oathmesh_middleware,
    get_oathmesh_context,
)
from oathmesh.middleware.decorators import (
    oathmesh_requires,
)
from oathmesh.middleware.types import OathMeshGraphQLConfig

# ─── GraphQL Schema ──────────────────────────────────────────────────────────


@strawberry.type
class User:
    """User represents an authenticated principal in the system."""

    id: strawberry.ID
    name: str
    
    @oathmesh_requires(require="action:read:user:email")
    async def email(self) -> Optional[str]:
        """Email field restricted to users with action:read:user:email scope."""
        return "user@example.com"


@strawberry.type
class Query:
    """GraphQL Query root type."""

    @strawberry.field
    async def current_user(self, info: Info) -> Optional[User]:
        """Get the currently authenticated user from the token.
        
        Returns the user if authenticated, None otherwise.
        """
        # Get verified claims from OathMesh middleware
        oathmesh = get_oathmesh_context(info.context)

        if not oathmesh or not oathmesh.verified:
            return None

        # In a real app, look up user from database using subject
        subject = oathmesh.claims.principal.subject
        print(f"✓ Authenticated as: {subject}")

        return User(
            id=strawberry.ID("1"),
            name="Test User",
        )


@strawberry.type
class Mutation:
    """GraphQL Mutation root type."""

    @strawberry.mutation
    async def update_user(self, info: Info, name: str) -> Optional[User]:
        """Update the current user's profile.
        
        Requires write permission.
        """
        oathmesh = get_oathmesh_context(info.context)

        if not oathmesh or not oathmesh.verified:
            raise Exception("Not authenticated")

        # Verify mutation permission
        if "action:write:user:profile" not in (oathmesh.claims.scope or []):
            raise Exception("Insufficient permissions to update user")

        return User(
            id=strawberry.ID("1"),
            name=name,
        )


# ─── Server Setup ────────────────────────────────────────────────────────────


async def get_context(request: Request) -> dict:
    """Create GraphQL context from FastAPI request.
    
    This context is passed to resolvers and middleware.
    """
    return {"request": request}


# Create the schema
schema = strawberry.Schema(query=Query, mutation=Mutation)

# Create FastAPI app
app = FastAPI(
    title="OathMesh GraphQL Example",
    description="GraphQL API with OathMesh authentication",
    version="1.0.0",
)


# ─── Middleware Configuration ────────────────────────────────────────────────

oathmesh_config = OathMeshGraphQLConfig(
    audience="https://api.example.com",
    trusted_issuers=[
        "https://issuer.oathmesh.tech",
        # Add your issuer URLs here
    ],
    queries_per_minute=100,
    mutations_per_minute=10,
    on_rate_limit_exceeded=lambda subject, op_type: print(
        f"⚠️  Rate limit exceeded for {subject} ({op_type})"
    ),
)


# ─── Routes ──────────────────────────────────────────────────────────────────

# Create GraphQL ASGI app with OathMesh middleware
graphql_app = GraphQL(
    schema,
    context_getter=get_context,  # type: ignore[arg-type]
    middleware=[create_oathmesh_middleware(oathmesh_config)],
)

# Mount GraphQL endpoint
app.mount("/graphql", graphql_app)


@app.get("/")
async def root():
    """API documentation."""
    return {
        "message": "OathMesh GraphQL API",
        "graphql_endpoint": "/graphql",
        "usage": {
            "query": {
                "without_token": 'curl -X POST http://localhost:8000/graphql -H "Content-Type: application/json" -d \'{"query": "{ currentUser { id name } }"}\'',
                "with_token": 'curl -X POST http://localhost:8000/graphql -H "Content-Type: application/json" -H "Authorization: OathMesh <token>" -d \'{"query": "{ currentUser { id name email } }"}\'',
            },
        },
    }


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "ok", "timestamp": datetime.utcnow().isoformat()}


if __name__ == "__main__":
    import uvicorn

    print("""
╔════════════════════════════════════════════════════════════════╗
║         OathMesh GraphQL Server - Strawberry                   ║
╚════════════════════════════════════════════════════════════════╝

🚀 Starting server on http://localhost:8000

📚 GraphQL Playground: http://localhost:8000/graphql

💡 Usage:

   # Query without token (should fail)
   curl -X POST http://localhost:8000/graphql \\
     -H "Content-Type: application/json" \\
     -d '{"query": "{ currentUser { id name } }"}'

   # Query with token
   curl -X POST http://localhost:8000/graphql \\
     -H "Content-Type: application/json" \\
     -H "Authorization: OathMesh <token>" \\
     -d '{"query": "{ currentUser { id name email } }"}'

⚠️  Note: This example uses a mock token.
   In production, tokens should be minted by OathMesh issuer.

Press CTRL+C to stop the server.
    """)

    uvicorn.run(app, host="0.0.0.0", port=8000)
