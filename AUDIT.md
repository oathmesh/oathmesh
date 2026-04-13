# Security Audit Preparation

> Checklist and documentation for third-party security audits.

## Audit Readiness

This document outlines the preparation for a third-party security audit of OathMesh.

## Architecture Overview

OathMesh is a machine-identity system with:
- **Go core**: Issuer, gateway, CLI (`cmd/oathmesh`)
- **Node.js SDK**: `sdk/node`
- **Python SDK**: `sdk/python`
- **Deployment**: Docker, Kubernetes

## Audit Scope

### In Scope

1. **Core Service** (`internal/`)
   - Token signing and verification
   - JWKS endpoint and caching
   - Policy evaluation engine

2. **SDKs**
   - Node.js SDK (`sdk/node/src/`)
   - Python SDK (`sdk/python/src/oathmesh/`)

3. **Infrastructure**
   - Docker deployment
   - Kubernetes manifests

### Out of Scope

- Upstream dependencies (unless known vulnerable)
- Demo scripts
- Documentation

## Security Controls

### Implemented Controls

1. **Cryptography**
   - Ed25519 for token signing (ES256 deprecated)
   - Key ID format: `issuer-key-YYYY-MM-{4-char-random-hex}`

2. **Request Binding**
   - `rqh` claim required for verification
   - Prevents token stealing attacks

3. **Replay Defense**
   - In-memory replay cache (default)
   - Redis-backed cache for multi-instance

4. **Policy Evaluation**
   - JSON-based policy with `match`/`allow` rules
   - Subject and action matching

5. **Network Security**
   - NetworkPolicy for K8s deployments
   - Issuer not publicly exposed

6. **TLS Enforcement**
   - HTTPS required for JWKS fetch in production
   - TLS 1.2+ required

## Key Files

### Core Implementation

- `internal/verify/verify.go` - Token verification logic
- `internal/sign/keyset.go` - Key management
- `internal/core/errors.go` - Error taxonomy

### SDKs

- `sdk/node/src/verify.ts` - Node.js verification
- `sdk/python/src/oathmesh/verify.py` - Python verification

### Configuration

- `.env.example` - Environment variables
- `docker-compose.yml` - Deployment config

## Running Audit-Prep

Run pre-audit checks:

```bash
make audit-prep
```

This will:
1. Run all tests
2. Run vulnerability scan (govulncheck)
3. Run linter
4. Generate audit report

## Vulnerability Disclosure

For vulnerability reports, contact: security@oathmesh.example

See: `security/vulnerability-disclosure.md`

## Previous Security Work

See: `CHANGELOG.md` for security-relevant changes.

## Audit Checklist

### Cryptography Review
- [ ] Ed25519 key generation and storage
- [ ] Key rotation procedures
- [ ] JWKS caching and TTL
- [ ] Key compromise response

### Token Security
- [ ] Request binding enforcement (`rqh`)
- [ ] Replay cache implementation
- [ ] TTL enforcement
- [ ] Algorithm selection (ES256 deprecation)

### SDK Review
- [ ] Node.js SDK verification
- [ ] Python SDK verification
- [ ] Cross-SDK conformance

### Deployment Security
- [ ] Kubernetes NetworkPolicy
- [ ] TLS configuration
- [ ] Secret management
- [ ] Audit logging

### Operational Security
- [ ] Key rotation procedures
- [ ] Incident response plan
- [ ] Monitoring and alerting
- [ ] Backup and recovery

## External Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [NIST Cryptographic Standards](https://csrc.nist.gov/projects/cryptographic-standards-and-guidelines)

## Audit Report Template

```
# OathMesh Security Audit Report

## Auditor
- Name:
- Organization:
- Date:

## Scope
- Components audited:
- Versions:

## Findings

### Critical
- [ ] ...

### High
- [ ] ...

### Medium
- [ ] ...

### Low
- [ ] ...

### Informational
- [ ] ...

## Recommendations

## Conclusion
```