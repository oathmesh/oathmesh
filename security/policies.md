# Security Policy

> OathMesh security policies and vulnerability disclosure guidelines.

## Supported Versions

| Version | Supported          |
| ------- | ----------------- |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

To report a security vulnerability:

1. **DO NOT** create a public GitHub issue
2. Email: security@oathmesh.example
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (optional)

## Response Timeline

- **Initial Response**: 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: 
  - Critical: 7 days
  - High: 14 days
  - Medium: 30 days
  - Low: Next release

## Scope

In scope:
- Token signing/verification bypass
- Key compromise
- Replay attack vulnerabilities
- Policy bypass
- JWKS injection
- Authentication bypass

Out of scope:
- Social engineering
- Physical attacks
- Denial of service (unless trivially exploitable)
- Issues in upstream dependencies

## Disclosure Policy

- Coordinated disclosure after fix is available
- Credit to reporters in release notes (opt-in)
- Public disclosure after 90 days or upon fix release

## Security Acknowledgments

Thank you to security researchers who have helped:
- (Add names here)

## Best Practices

When deploying OathMesh:

1. **Key Management**
   - Use strong key generation: `openssl genpkey -algorithm Ed25519`
   - Rotate keys regularly (see: `docs/deployment/key-rotation.md`)
   - Never commit keys to version control

2. **Network Security**
   - Use TLS in production (see: `docs/deployment/tls.md`)
   - Apply NetworkPolicy in Kubernetes
   - Don't expose issuer publicly

3. **Monitoring**
   - Enable audit logging
   - Monitor for unusual patterns
   - Set up alerts for key operations

4. **Updates**
   - Keep OathMesh updated
   - Monitor security advisories
   - Subscribe to releases