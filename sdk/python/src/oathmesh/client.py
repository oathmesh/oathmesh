import json
import time
from typing import Optional, List, Dict, Any
import urllib.request
import urllib.error

from .errors import OathMeshError


class OathMeshClient:
    """OathMeshClient auto-manages token minting and caching securely."""

    def __init__(self, issuer: str, api_key: str):
        self.issuer = issuer.rstrip("/")
        self.api_key = api_key
        self._cache: Dict[str, Dict[str, Any]] = {}

    def mint(
        self,
        sub: str,
        aud: str,
        act: str,
        ttl_hint: Optional[int] = None,
        nbf_hint: Optional[int] = None,
        scope: Optional[List[str]] = None,
        reason: Optional[str] = None,
        env: Optional[str] = None,
        rqh: Optional[str] = None,
    ) -> str:
        """Mint a new token, caching it automatically until near its TTL."""
        
        req_dict = {
            "sub": sub,
            "aud": aud,
            "act": act,
        }
        if ttl_hint is not None:
            req_dict["ttl_hint"] = ttl_hint
        if nbf_hint is not None:
            req_dict["nbf_hint"] = nbf_hint
        if scope is not None:
            req_dict["scope"] = scope
        if reason is not None:
            req_dict["reason"] = reason
        if env is not None:
            req_dict["env"] = env
        if rqh is not None:
            req_dict["rqh"] = rqh

        cache_key = json.dumps(req_dict, sort_keys=True)
        cached = self._cache.get(cache_key)

        # Auto-refresh: return cache if valid for at least 15 more secs
        if cached and cached["expires_at"] > time.time() + 15:
            return cached["token"]

        url = f"{self.issuer}/v1/token"
        req_data = json.dumps(req_dict).encode("utf-8")
        
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}",
        }
        
        req = urllib.request.Request(url, data=req_data, headers=headers, method="POST")
        
        try:
            with urllib.request.urlopen(req) as response:
                if response.status != 200:
                    raise OathMeshError("verification_failed", f"Failed to mint token: {response.status}")
                res_data = json.loads(response.read().decode("utf-8"))
        except urllib.error.HTTPError as e:
            raise OathMeshError("verification_failed", f"Failed to mint token: HTTP {e.code}")
        except Exception as e:
            raise OathMeshError("verification_failed", f"Failed to mint token: {str(e)}")

        token = res_data.get("token")
        expires_in = res_data.get("expires_in", 0)
        
        self._cache[cache_key] = {
            "token": token,
            "expires_at": time.time() + expires_in
        }
        
        return token
