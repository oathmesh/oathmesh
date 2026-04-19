"""OathMesh GraphQL Middleware Types

Type definitions and data structures for OathMesh GraphQL middleware integration.
"""

from __future__ import annotations
from dataclasses import dataclass, field
from typing import Dict, Optional, Any, Callable
from datetime import datetime
from oathmesh.types import VerifiedCallerContext


@dataclass
class OathMeshGraphQLContext:
    """OathMesh context extension for GraphQL operations.
    
    Attached to the GraphQL context during middleware execution.
    """
    
    claims: VerifiedCallerContext
    """Verified caller claims from token verification."""
    
    verified: bool
    """Whether the token was successfully verified."""
    
    rate_limit: Optional[Dict[str, Any]] = None
    """Rate limit status."""


@dataclass
class OathMeshGraphQLConfig:
    """Configuration for OathMesh GraphQL middleware.
    
    Attributes:
        audience: The audience URL this API expects.
        trusted_issuers: List of trusted issuer URLs.
        require_request_binding: Enforce rqh claim for mutations.
        replay_cache: Optional replay cache implementation.
        policy_evaluator: Optional policy evaluator for authorization.
        queries_per_minute: Rate limit for queries (default: 100).
        mutations_per_minute: Rate limit for mutations (default: 10).
        on_rate_limit_exceeded: Callback when rate limit is exceeded.
    """
    
    audience: str
    trusted_issuers: list[str]
    require_request_binding: bool = False
    replay_cache: Optional[Any] = None
    policy_evaluator: Optional[Any] = None
    queries_per_minute: int = 100
    mutations_per_minute: int = 10
    on_rate_limit_exceeded: Optional[Callable[[str, str], None]] = None


@dataclass
class RateLimitBucket:
    """Rate limit bucket for a subject.
    
    Tracks request counts per operation type within a time window.
    """
    
    request_counts: Dict[str, int] = field(default_factory=dict)
    window_start: float = field(default_factory=lambda: __import__('time').time())
