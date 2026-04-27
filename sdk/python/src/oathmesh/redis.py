from typing import Any
from .types import RevocationList

class RedisRevocationCache(RevocationList):
    """Redis-backed revocation cache implementation."""
    
    def __init__(self, client: Any, prefix: str = "om:rev:"):
        """
        Args:
            client: A redis-py client instance.
            prefix: Key prefix for redis.
        """
        self._client = client
        self._prefix = prefix

    def is_revoked(self, subject: str) -> bool:
        return self._client.get(f"{self._prefix}{subject}") is not None
