"""OathMesh types mirroring the Go VerifiedCallerContext."""

from __future__ import annotations
from dataclasses import dataclass, field
from typing import List, Optional


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
