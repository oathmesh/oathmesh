"""OathMesh types mirroring the Go VerifiedCallerContext."""

from __future__ import annotations
from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any


@dataclass(frozen=True)
class Principal:
    """Authenticated identity of the caller."""
    issuer: str
    subject: str


@dataclass(frozen=True)
class Source:
    """Source provenance — where the call originated."""
    type: str
    repo: Optional[str] = None
    workflow: Optional[str] = None
    run_id: Optional[str] = None
    sha: Optional[str] = None


@dataclass(frozen=True)
class VerifiedCallerContext:
    """The verified identity after successful token verification.

    This is the structured output of a successful verification.
    Receivers use this to make authorization decisions.
    """
    principal: Principal
    action: str
    token_id: str
    environment: str = ""
    scope: List[str] = field(default_factory=list)
    reason: Optional[str] = None
    source: Optional[Source] = None


class ReplayCache:
    """Replay cache interface for preventing token reuse attacks.
    
    Implementations can be in-memory (for single-instance) or Redis (for multi-instance).
    """
    
    def check(self, jti: str) -> bool:
        """Check if a token JTI has been seen before.
        
        Returns True if the token has been replayed (already used).
        Returns False if this is the first time seeing this JTI.
        """
        raise NotImplementedError
    
    def add(self, jti: str) -> None:
        """Record a token JTI as seen.
        
        Should be called after successful verification.
        """
        raise NotImplementedError


class InMemoryReplayCache(ReplayCache):
    """In-memory replay cache implementation for development/single-instance.
    
    Uses a dict with TTL to automatically expire old entries.
    """
    
    def __init__(self, default_ttl: int = 300):
        self._cache: Dict[str, int] = {}
        self._default_ttl = default_ttl
    
    def check(self, jti: str) -> bool:
        expires_at = self._cache.get(jti)
        if expires_at is None:
            return False
        if datetime.now().timestamp() > expires_at:
            del self._cache[jti]
            return False
        return True
    
    def add(self, jti: str) -> None:
        self._cache[jti] = datetime.now().timestamp() + self._default_ttl


from datetime import datetime


@dataclass
class PolicyInput:
    """Policy input for evaluation."""
    iss: str
    sub: str
    aud: str
    act: str
    scope: Optional[List[str]] = None
    env: Optional[str] = None


@dataclass
class PolicyDecision:
    """Policy decision result."""
    outcome: str  # 'allow' or 'deny'
    rule_name: Optional[str] = None
    deny_reason: Optional[str] = None


class PolicyEvaluator:
    """Policy evaluator interface.
    
    Implementations evaluate token claims against policy rules.
    """
    
    def evaluate(self, input: PolicyInput) -> PolicyDecision:
        raise NotImplementedError


@dataclass
class JsonPolicyRule:
    """JSON policy rule format."""
    match: Optional[Dict[str, Any]] = None
    allow: bool = True
    rule_name: Optional[str] = None
    deny_reason: Optional[str] = None


@dataclass
class JsonPolicyDocument:
    """JSON policy document format."""
    rules: List[JsonPolicyRule] = field(default_factory=list)


class JsonPolicyEvaluator(PolicyEvaluator):
    """JSON policy evaluator that loads and evaluates simple JSON policies."""
    
    def __init__(self, policy: JsonPolicyDocument):
        self._policy = policy
    
    def evaluate(self, input: PolicyInput) -> PolicyDecision:
        for rule in self._policy.rules:
            if self._matches_rule(input, rule):
                if rule.allow:
                    return PolicyDecision(outcome='allow', rule_name=rule.rule_name)
                else:
                    return PolicyDecision(outcome='deny', rule_name=rule.rule_name, deny_reason=rule.deny_reason)
        # Default deny if no rules match
        return PolicyDecision(outcome='deny', deny_reason='no matching policy rule')
    
    def _matches_rule(self, input: PolicyInput, rule: JsonPolicyRule) -> bool:
        if not rule.match:
            return False
        match = rule.match
        
        if 'sub' in match and not self._match_pattern(input.sub, match['sub']):
            return False
        if 'aud' in match and not self._match_pattern(input.aud, match['aud']):
            return False
        if 'act' in match and not self._match_pattern(input.act, match['act']):
            return False
        if 'env' in match and input.env != match['env']:
            return False
        
        return True
    
    def _match_pattern(self, value: str, pattern: str) -> bool:
        if '*' in pattern:
            import re
            regex = re.compile('^' + pattern.replace('*', '.*') + '$')
            return regex.match(value) is not None
        return value == pattern
