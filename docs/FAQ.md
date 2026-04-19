# OathMesh FAQ

## 1) What problem does OathMesh solve?
OathMesh secures machine-to-machine calls with short-lived signed identity tokens, replacing static API keys and long-lived shared secrets.  
See: [overview.md](overview.md)

## 2) Is OathMesh for end-user login?
No. OathMesh is for service/workload identity on backend calls, not browser login flows or user sessions.  
See: [concepts.md](concepts.md)

## 3) What does the issuer do?
The issuer validates token requests and signs tokens with Ed25519 keys.  
See: [protocol/token-format.md](protocol/token-format.md)

## 4) What does the receiver do?
The receiver verifies token structure, signature, claims, replay safety, and policy before allowing access.  
See: [protocol/verification-rules.md](protocol/verification-rules.md)

## 5) How short-lived are tokens?
By design, tokens are short-lived (with strict max TTL guidance) to reduce blast radius and credential reuse risk.  
See: [overview.md](overview.md)

## 6) How is replay prevented?
Each token includes a unique `jti`; receivers track and reject reused IDs inside the validity window.  
See: [security/replay-defense.md](security/replay-defense.md)

## 7) What claims matter most?
`sub` (who), `aud` (target), `act` (intent), `iat`/`exp` (time), and `jti` (uniqueness).  
See: [protocol/claim-reference.md](protocol/claim-reference.md)

## 8) How is authorization decided?
OathMesh verifies identity; your service policy decides authorization. Default-deny is recommended.  
See: [policies/overview.md](policies/overview.md)

## 9) Can I use OathMesh with existing APIs?
Yes. Start at the edge (gateway/middleware) and migrate endpoint by endpoint.  
See: [migration/replace-api-key.md](migration/replace-api-key.md)

## 10) What frameworks are supported?
Examples and quickstarts exist for Go (chi), Express, Next.js, and FastAPI.  
See: [quickstarts/](quickstarts/)

## 11) Can CI/CD pipelines use OathMesh?
Yes. CI jobs can mint scoped short-lived tokens for internal deployments and API calls.  
See: [quickstarts/github-actions-to-internal-api.md](quickstarts/github-actions-to-internal-api.md)

## 12) How should I manage signing keys?
Use environment separation, strict access controls, and planned rotation.  
See: [security/key-management.md](security/key-management.md)

## 13) How often should keys rotate?
Use a regular schedule plus emergency rotation procedures.  
See: [deployment/key-rotation.md](deployment/key-rotation.md) and [operations/key-rotation.md](operations/key-rotation.md)

## 14) What should we log?
Log security decisions and context, not secrets or full token payloads.  
See: [security/logging-guidance.md](security/logging-guidance.md)

## 15) Is OathMesh compliance-ready?
OathMesh supports compliance operations, but your final posture depends on deployment and process controls.  
See: [security/soc2-compliance.md](security/soc2-compliance.md), [enterprise/compliance.md](enterprise/compliance.md)

## 16) Where do I start in production?
Use the production checklist, then wire monitoring and incident runbooks before broad rollout.  
See: [operations/production-checklist.md](operations/production-checklist.md), [operations/on-call-runbook.md](operations/on-call-runbook.md)

## 17) Where can I ask for help or contribute?
Use support and contribution docs, then join community discussions through issues/PRs.  
See: [COMMUNITY.md](COMMUNITY.md), [../SUPPORT.md](../SUPPORT.md), [../CONTRIBUTING.md](../CONTRIBUTING.md)
