# OathMesh Overview

<p align="center">
  <b>🔐 Short-lived, signed identity tokens for machine-to-machine calls.</b>
</p>

<p align="center">
  <a href="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml">
    <img src="https://github.com/oathmesh/oathmesh/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/releases">
    <img src="https://img.shields.io/github/v/release/oathmesh/oathmesh" alt="Release">
  </a>
  <a href="https://github.com/oathmesh/oathmesh/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/oathmesh/oathmesh" alt="License">
  </a>
</p>

---

> 🆕 **New here?** Start with the [Quick Start](../README.md#-quick-start) in the main README.

## What Is OathMesh

OathMesh is a micro-protocol and developer platform that gives every machine-to-machine call a short-lived, cryptographically signed identity. It replaces shared secrets—API keys, static tokens, long-lived credentials—with scoped, verifiable, auditable call assertions.

## When to Use OathMesh

Use OathMesh when:

- **Service A calls Service B** and Service B needs to know *who* is calling, *what* they want to do, and *whether they are allowed*—without sharing a static secret.
- **CI/CD jobs** (GitHub Actions, GitLab CI) need to authenticate against internal APIs without embedding long-lived tokens in environment variables.
- **AI agents and bots** make tool calls and need scoped, time-limited credentials that expire automatically.
- **Internal tooling** calls production APIs and you need an audit trail of who called what, when, and why.

## When NOT to Use OathMesh

OathMesh is **not**:

- A **user authentication** system. It does not handle browser logins, OAuth flows, or session management for humans.
- A **service mesh** or data plane. It authenticates individual calls; it does not route traffic or manage service discovery.
- A **replacement for cloud IAM**. Use your cloud provider's IAM for infrastructure-level permissions. Use OathMesh for application-level call identity.
- A **replacement for SPIFFE/SPIRE**. OathMesh can run alongside SPIFFE. SPIFFE identifies workloads; OathMesh identifies individual calls from those workloads.

## Core Doctrine

> **"OathMesh authenticates the caller. The receiver authorizes the request."**

OathMesh proves *who* is making a call, *what action* they want to perform, and *where* the call originated. The receiving service decides whether to allow it based on its own policy. OathMesh never makes authorization decisions on the caller's behalf.

## How It Works

1. **The Caller** requests a short-lived Oath Token from the **Issuer**, declaring who they are (`sub`), what they want (`act`), and who they're calling (`aud`).
2. **The Issuer** validates the request, signs the token with Ed25519, enforces a maximum TTL of 300 seconds, and returns it.
3. **The Caller** attaches the token to the outgoing request: `Authorization: OathMesh <token>`.
4. **The Receiver** (or the OathMesh Gateway sitting in front of it) verifies the token through 14 mandatory steps: structure, signature, issuer trust, expiry, audience, required claims, replay detection, and policy evaluation.
5. **If verification passes**: the receiver gets a `VerifiedCallerContext` containing the caller's identity, action, source provenance, and environment—everything needed to make an authorization decision.
6. **If verification fails**: the receiver gets a structured error with a machine-readable code, human-readable message, and a fix instruction.

## Key Properties

| Property | Guarantee |
|---|---|
| **Short-lived** | Every token expires in ≤ 300 seconds. No refresh tokens. No long-lived credentials. |
| **Signed** | Ed25519 signatures. No symmetric keys. No `alg: "none"`. |
| **Scoped** | Every token declares exactly one audience and one action. |
| **Auditable** | Every verification emits an NDJSON audit event—allow and deny. |
| **Replay-safe** | Every token has a unique `jti`. Replay cache rejects duplicates within the TTL window. |
| **Policy-driven** | Pkl-based policy rules evaluated at verification time. Default deny. |
