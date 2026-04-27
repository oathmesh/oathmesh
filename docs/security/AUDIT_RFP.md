# Request for Proposal (RFP): OathMesh Security Audit

**Target Audience:** Trail of Bits, Cure53, NCC Group, or equivalent high-tier security evaluation firms.
**Project:** OathMesh - Polyglot Zero-Trust Machine Identity Protocol
**Phase:** Pre-1.0 Production Readiness Evaluation
**Date:** 2026-04-27

---

## 1. Executive Summary

OathMesh is a high-performance, polyglot protocol designed for verifiable machine-to-machine (M2M) identity and zero-trust authorization in distributed environments. It operates primarily via EdDSA-signed JSON Web Tokens (JWTs) representing machine identities.

We are seeking a comprehensive security audit of our v0.3.0 pre-release to mathematically and practically verify the resilience of our cryptographic pipeline, policy evaluation sandboxing, and polyglot SDK parity.

## 2. Architecture & Components

OathMesh consists of three major components:
1.  **OathMesh Core (Go):** The reference implementation containing the 14-step zero-trust verification pipeline, JWKS caching, and the Pkl policy execution engine. This acts as both the CLI (`oathmesh`), Gateway proxy, and Envoy/Kong external authorization server.
2.  **Polyglot SDKs (Node.js & Python):** Lightweight, idiomatic libraries that mirror the exact 14-step verification semantics of the Go core. They perform local, decentralized cryptographic verification to eliminate network hops.
3.  **Policy Engine (Pkl):** OathMesh delegates fine-grained attribute-based access control (ABAC) to Apple's Pkl language, executing policies in a restricted, sandboxed environment.

## 3. Scope of Engagement

Please refer to `docs/security/AUDIT_SCOPE.md` for the explicit boundaries of the evaluation. High-priority areas include:

### 3.1. Cryptographic Pipeline (14-Step Verification)
- **Token Parsing & Validation:** Ensuring `token_malformed` rejections are robust against maliciously crafted JWTs, `alg:none` attacks, and symmetric/asymmetric confusion.
- **Clock Skew:** Validating the mathematical correctness of our 30-second `ClockSkewLeeway` applied to `exp`, `nbf`, and `iat` claims.
- **Fail-Closed Caches:** Validating that our `ReplayCache` and `RevocationList` (InMemory and Redis-backed) correctly fail-closed (returning `verification_failed` HTTP 401/gRPC 16) during network partitions or Redis outages.

### 3.2. Pkl Sandboxing & Execution
OathMesh utilizes Pkl via OS execution (`exec.Command`). We have implemented strict sandboxing flags (`--allowed-modules="pkl:*"` and `--allowed-resources="env:*,prop:*,file://<dir>/"`).
- **Objective:** Attempt to achieve Server-Side Request Forgery (SSRF), Local File Inclusion (LFI), or Remote Code Execution (RCE) by crafting a malicious `.pkl` policy or bypassing the enforced `file://<dir>/` chroot-like isolation.

### 3.3. Key Management
- Validating the `oathmesh keygen` process securely generates Ed25519 pairs with POSIX `0600` permissions.
- Verifying the deprecation and mitigation of plaintext environment variables.

### 3.4. SDK Parity
- Evaluating the TypeScript (Node.js) and Python SDKs against the Go reference implementation to guarantee identical failure modes for edge-case tokens (0% mismatch requirement).

## 4. Engagement Details

- **Timeline:** We aim to schedule the audit window within Q3 2026.
- **Deliverables:** A formal vulnerability report, an executive summary, and remediation validation upon our application of patches.
- **Access:** Full source code access, automated test suites (including our cross-framework Bash conformance suite), and Kubernetes deployment topologies will be provided.

## 5. Threat Model

Our explicit threat model assumes:
1.  **Compromised Upstream:** An attacker who has compromised a service within the mesh cannot forge a valid OathMesh token for lateral movement without the private Ed25519 key.
2.  **Stale/Revoked Credentials:** Tokens are strictly bound by a 300s maximum TTL and immediate invalidation via Redis revocation caches.
3.  **Replay Attacks:** Identical tokens intercepted in transit cannot be reused within their valid TTL window due to strict JTI tracking.

Please review this RFP and our accompanying documentation. We look forward to your proposal and scoping discussion.
