---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
review_by: "2026-07-05"
purpose: "Key components, services, people, and systems the AI must always recognize"
---

# OathMesh Entity Memory

## Internal Components

| Entity | Type | Description | Key Files |
|---|---|---|---|
| Issuer Service | Service | Mints and signs Oath Tokens, serves JWKS and metadata | `/issuer/` (planned) |
| Verifier Middleware | Library | Validates tokens in receiver applications | `/sdk-node/`, `/sdk-python/` (planned) |
| Policy Engine | Library | Evaluates YAML policy rules against Verified Caller Context | Embedded in verifier middleware |
| Gateway | Service | Reverse proxy with built-in token verification | `/gateway/` (planned) |
| CLI | Tool | `oathmesh` command-line interface | `/issuer/cmd/oathmesh/` (planned) |
| Audit Logger | Library | Emits structured JSON audit events | Embedded in verifier middleware and gateway |

## External Systems

| Entity | Type | Relationship to OathMesh | Key Details |
|---|---|---|---|
| GitHub Actions OIDC | Identity Provider | Primary bootstrap identity source (golden path) | Issuer URL: `https://token.actions.githubusercontent.com` |
| GitHub JWKS | Key Server | Provides public keys for GitHub OIDC token verification | URL: `https://token.actions.githubusercontent.com/.well-known/jwks` |
| Kubernetes API Server | Identity Provider | Secondary bootstrap source for cluster workloads | Uses ServiceAccount OIDC discovery |
| Cloud KMS (AWS/GCP/Azure) | Key Management | Stores and manages issuer signing keys in production | Accessed via cloud SDK |
| Docker / Container Runtime | Infrastructure | Runs issuer, gateway, and example services | Docker Compose for dev, Kubernetes for prod |

## Standards & Specifications

| Entity | Type | Relevance |
|---|---|---|
| RFC 7519 (JWT) | Standard | Token format foundation |
| RFC 7515 (JWS) | Standard | Token signing format |
| RFC 7517 (JWK/JWKS) | Standard | Key format and distribution |
| RFC 8037 (EdDSA for JOSE) | Standard | Preferred signing algorithm |
| RFC 8615 (Well-Known URIs) | Standard | Metadata endpoint pattern |
| SPIFFE | Standard | Related workload identity standard — OathMesh is not SPIFFE, but may accept SPIFFE IDs as subjects |

## People & Roles

| Entity | Role | Key Responsibilities |
|---|---|---|
| Founder | Product owner, protocol designer | Final authority on spec, ADRs, security redlines, go-to-market |

## Competitive Landscape

| Entity | Type | Relationship |
|---|---|---|
| SPIFFE/SPIRE | Competitor (enterprise) | Full workload identity framework — heavier, CNCF-backed. OathMesh is lighter, developer-first. |
| Aembit | Competitor (enterprise) | Workload IAM platform — broader scope. OathMesh is narrower, open-source-first. |
| Teleport | Competitor (infrastructure) | Infrastructure access platform — different focus (SSH, DB, K8s access). Some overlap in machine identity. |
| Shared API Keys | Incumbent pattern | The primary thing OathMesh replaces. Every engineering team uses these today. |
| GitHub Actions OIDC + Cloud IAM | Partial solution | Works for CI-to-cloud, but doesn't cover internal APIs or agent-to-API calls. OathMesh fills this gap. |

## Key URLs (Planned)

| Entity | URL | Purpose |
|---|---|---|
| GitHub Repository | `https://github.com/oathmesh/oathmesh` | Open source codebase (planned) |
| Documentation Site | `https://docs.oathmesh.dev` | Developer documentation (planned) |
| Landing Page | `https://oathmesh.dev` | Marketing and product page (planned) |
| Dev Issuer | `https://issuer.dev.oathmesh.dev` | Hosted development issuer (Phase 2) |
| npm Package | `@oathmesh/node` | Node.js SDK (planned) |
| PyPI Package | `oathmesh` | Python SDK (planned) |
