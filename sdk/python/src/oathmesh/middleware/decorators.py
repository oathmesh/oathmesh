"""OathMesh Field Authorization Decorators

Decorators for field-level access control in GraphQL resolvers.
"""

from __future__ import annotations
import asyncio
import inspect
from functools import wraps
from typing import Callable, Any, Optional, TypeVar, Union

T = TypeVar('T')

# Rate limited context
_field_permissions_cache: dict[str, set[str]] = {}


def check_field_permission(context: Any, requirement: str) -> bool:
    """Check if a field requirement is satisfied.
    
    Requirement formats:
    - "action:read:user:email" — exact scope match
    - "role:admin" — exact scope match
    - "action:*" — wildcard match (matches any scope starting with "action:")
    
    Args:
        context: GraphQL context (must contain oathmesh property)
        requirement: Permission requirement string
        
    Returns:
        True if permission is granted, False otherwise
    """
    from .graphql import get_oathmesh_context
    
    oathmesh = get_oathmesh_context(context)
    
    if not oathmesh or not oathmesh.verified:
        return False
    
    claims = oathmesh.claims
    if not claims.scope:
        return False
    
    # Exact scope match
    if requirement in claims.scope:
        return True
    
    # Wildcard match (e.g., "action:*")
    if '*' in requirement:
        # Handle "action:*" to match "action:read", "action:write", etc.
        prefix = requirement.rstrip('*')
        return any(s.startswith(prefix) for s in claims.scope)
    
    # Also check if any scope is a wildcard that matches requirement
    # e.g., scope "action:*" matches requirement "action:read"
    for scope in claims.scope:
        if '*' in scope:
            prefix = scope.rstrip('*')
            if requirement.startswith(prefix):
                return True
    
    return False


def oathmesh_requires(require: str) -> Callable[[Callable[..., T]], Callable[..., T]]:
    """Decorator for field-level authorization in GraphQL resolvers.
    
    Use this to protect field resolvers with OathMesh permissions.
    
    Args:
        require: Permission requirement string (e.g., "action:read:user:email")
        
    Returns:
        Decorated resolver function
        
    Example:
        ```python
        import strawberry
        from oathmesh.middleware.decorators import oathmesh_requires
        
        @strawberry.type
        class User:
            id: strawberry.ID
            name: str
            
            @oathmesh_requires(require="action:read:user:email")
            async def email(self, info: Info) -> Optional[str]:
                return self._email
        ```
    """
    
    def decorator(func: Callable[..., T]) -> Callable[..., T]:
        is_async = asyncio.iscoroutinefunction(func)
        
        if is_async:
            @wraps(func)
            async def async_wrapper(*args: Any, **kwargs: Any) -> Any:
                # Extract context from arguments (could be first arg or in kwargs)
                context = None
                # Check first argument
                if args and hasattr(args[0], 'context'):
                    context = args[0].context
                # Check kwargs for info
                elif 'info' in kwargs and hasattr(kwargs['info'], 'context'):
                    context = kwargs['info'].context
                
                if not check_field_permission(context, require):
                    return None
                
                return await func(*args, **kwargs)
            
            return async_wrapper  # type: ignore
        else:
            @wraps(func)
            def sync_wrapper(*args: Any, **kwargs: Any) -> Any:
                # Extract context from arguments (could be first arg or in kwargs)
                context = None
                # Check first argument
                if args and hasattr(args[0], 'context'):
                    context = args[0].context
                # Check kwargs for info
                elif 'info' in kwargs and hasattr(kwargs['info'], 'context'):
                    context = kwargs['info'].context
                
                if not check_field_permission(context, require):
                    return None
                
                return func(*args, **kwargs)
            
            return sync_wrapper  # type: ignore
    
    return decorator


def requires_authenticated() -> Callable[[Callable[..., T]], Callable[..., T]]:
    """Decorator to require authentication for a resolver.
    
    Returns None if user is not authenticated.
    
    Example:
        ```python
        @strawberry.field
        @requires_authenticated()
        async def my_field(self, info: Info) -> str:
            return "secret"
        ```
    """
    
    def decorator(func: Callable[..., T]) -> Callable[..., T]:
        is_async = asyncio.iscoroutinefunction(func)
        
        if is_async:
            @wraps(func)
            async def async_wrapper(*args: Any, **kwargs: Any) -> Any:
                context = None
                # Check first argument
                if args and hasattr(args[0], 'context'):
                    context = args[0].context
                # Check kwargs for info
                elif 'info' in kwargs and hasattr(kwargs['info'], 'context'):
                    context = kwargs['info'].context
                
                from .graphql import get_oathmesh_context
                oathmesh = get_oathmesh_context(context)
                
                if not oathmesh or not oathmesh.verified:
                    return None
                
                return await func(*args, **kwargs)
            
            return async_wrapper  # type: ignore
        else:
            @wraps(func)
            def sync_wrapper(*args: Any, **kwargs: Any) -> Any:
                context = None
                # Check first argument
                if args and hasattr(args[0], 'context'):
                    context = args[0].context
                # Check kwargs for info
                elif 'info' in kwargs and hasattr(kwargs['info'], 'context'):
                    context = kwargs['info'].context
                
                from .graphql import get_oathmesh_context
                oathmesh = get_oathmesh_context(context)
                
                if not oathmesh or not oathmesh.verified:
                    return None
                
                return func(*args, **kwargs)
            
            return sync_wrapper  # type: ignore
    
    return decorator


class PermissionDenied(Exception):
    """Raised when a field permission check fails.
    
    Use this to explicitly deny access with a custom error message.
    """
    
    def __init__(self, message: str = "Permission denied"):
        self.message = message
        super().__init__(message)
