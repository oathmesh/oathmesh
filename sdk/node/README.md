---
title: Node SDK
description: OathMesh token verification SDK for Node.js and TypeScript with Express and Next.js integrations.
keywords: [oathmesh, nodejs, typescript, express, nextjs, jwt, machine-identity]
audience: application-developers, platform-engineers, enterprise-teams
---

[Docs Index](../../docs/INDEX.md)

# @oathmesh/sdk

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

OathMesh token verification for Node.js and TypeScript.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Framework Integrations](#framework-integrations)
- [Configuration](#configuration)
- [Error Handling](#error-handling)
- [Troubleshooting](#troubleshooting)
- [Security Notes](#security-notes)
- [Production Tips](#production-tips)
- [Related Docs](#related-docs)

## Installation

```bash
npm install @oathmesh/sdk
# or
yarn add @oathmesh/sdk
# or
pnpm add @oathmesh/sdk
```

Requirements:
- Node.js 18+
- Express 4+ (optional)
- Next.js 13+ (optional, for Next.js adapters)

## Quick Start

```typescript
import express from 'express';
import { verifyToken } from '@oathmesh/sdk';

const app = express();

app.use(verifyToken({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.tech'],
}));

app.get('/inventory', (req, res) => {
  res.json({ subject: req.oathmeshContext?.principal.subject });
});
```

Send tokens as:

```http
Authorization: OathMesh <token>
```

## Framework Integrations

### Express

Use `verifyToken(...)` middleware from `@oathmesh/sdk`.

### Next.js App Router

```typescript
import { withOathMesh } from '@oathmesh/sdk/next';

const oathmesh = withOathMesh({
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.tech'],
});
```

### Next.js Pages Router

Use `withOathMeshApi(...)` from `@oathmesh/sdk/next`.

### Next.js Edge

Use `createEdgeVerifier(...)` from `@oathmesh/sdk/next`.

## Configuration

```typescript
import { verifyOathToken, extractToken } from '@oathmesh/sdk';

const token = extractToken(headers.authorization);
const caller = await verifyOathToken(token!, {
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.oathmesh.tech'],
  onVerified: (ctx) => metrics.increment('oathmesh.allow', { sub: ctx.principal.subject }),
  onDenied: (err) => logger.warn('oathmesh.denied', { code: err.code, step: err.step }),
});
```

Verifier semantics:
- Canonical and required header for verification is `Authorization: OathMesh <token>`.
- `extractToken` returns `null` for non-OathMesh schemes.
- If upstream sends `Bearer`, translate it to `OathMesh` before calling verifier APIs.
- Verifier behavior is aligned with canonical Go step semantics for conformance-critical checks (including `alg=none` rejection, subject format validation, binding-required semantics, and future-`iat` rejection).
- Parity is behavioral across language runtimes; implementations are intentionally language-native, not byte-identical.
- `revocationList` remains optional. Cross-SDK conformance currently marks Node revocation behavior as SKIP/N/A.

## Error Handling

Verification failures use `OathMeshError` and map to a stable 401 payload shape.

```typescript
import { OathMeshError } from '@oathmesh/sdk';

try {
  // verify...
} catch (err) {
  if (err instanceof OathMeshError) {
    console.error(err.code, err.message, err.fix, err.step);
  }
}
```

Common codes: `claim_missing:token`, `issuer_untrusted`, `audience_mismatch`, `signature_invalid`, `token_expired`.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `claim_missing:token` | Missing/invalid header | Send `Authorization: OathMesh <token>` |
| `issuer_untrusted` | `iss` not in `trustedIssuers` | Add exact issuer URL |
| `audience_mismatch` | Token `aud` differs from `audience` | Mint token with matching `aud` |
| `token_expired` | TTL elapsed / clock skew | Mint fresh token and sync clocks |
| `signature_invalid` | JWKS/issuer mismatch | Check issuer URL and JWKS reachability |

## Security Notes

- Verify only tokens from trusted issuers you control.
- Never log raw tokens or API keys.
- Keep issuer TLS and JWKS endpoints reachable and authenticated.
- Treat all caller claims as untrusted until verification succeeds.

## Production Tips

- Reuse verifier config and HTTP/JWKS clients to reduce latency.
- Record `error code` and `step` in logs for fast triage.
- Use short token TTLs and one token per operation where possible.
- Add metrics for allow/deny rates and failure categories.

## Related Docs

- [Getting Started](../../docs/GETTING_STARTED.md)
- [Troubleshooting Guide](../../docs/TROUBLESHOOTING.md)
- [Community](../../docs/COMMUNITY.md)
- [Enterprise Guide](../../docs/enterprise/README.md)
