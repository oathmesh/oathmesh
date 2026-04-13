# Security Policy

<p align="center">
  <img src="assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  OathMesh is built for security-critical systems. We take security seriously and appreciate responsible disclosure.
</p>

---

## Supported Versions

| Version | Supported          |
|---------|-------------------|
| 1.0.x   | ✅ Active         |
| < 1.0   | ❌ Not supported  |

## Scope

### What's Covered

- Token signing and verification (Ed25519)
- 14-step verification pipeline
- Policy engine evaluation
- Replay attack prevention
- JWKS fetching and caching
- SDK middleware (Go, Node, Python)

### What's NOT Covered

- User authentication / OAuth flows
- Service mesh routing
- Cloud IAM integration
- Your upstream service implementation

## Reporting a Vulnerability

### ⚠️ Do NOT open a public issue for security vulnerabilities.

**Responsible Disclosure:**

1. **Email:** security@oathmesh.dev
2. **GitHub:** Use [Private Vulnerability Reporting](https://github.com/oathmesh/oathmesh/security/advisories/new)

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes (optional)

### Our Commitment

- **Acknowledge:** Within 48 hours
- **Timeline:** Fix within 30 days (critical: faster)
- **Disclosure:** Coordinated public disclosure after fix

## Security Features

| Feature | Implementation |
|---------|---------------|
| Token expiry | Max 300 seconds TTL |
| Cryptography | Ed25519 (EdDSA) only |
| Replay protection | Unique `jti` + cache |
| Default deny | Policy must explicitly allow |
| Audit logging | Every verify attempt logged |

## Security Best Practices

When using OathMesh:

1. **Never log full tokens** — Log `jti` + claim summary only
2. **Never expose private keys** — Load from env, rotate regularly
3. **Use gateway mode** — Strips tokens, injects context headers
4. **Enable audit logging** — Every allow/deny matters
5. **Rotate keys** — Regular key rotation schedule

## Third-Party Security

- ✅ Go stdlib `crypto/ed25519`
- ✅ Node.js `jose` library
- ✅ Python `PyJWT` with cryptography

No external dependencies on critical security paths.

---

<p align="center">
  <sub>Thank you for helping keep OathMesh secure! 🛡️</sub>
</p>
