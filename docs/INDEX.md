# 📚 OathMesh Documentation Index

Welcome to OathMesh docs. Start with the canonical route, then jump to a category.

---

## 🚦 Canonical Route (Start Here)

Use this canonical Start Here flow referenced from README and GETTING_STARTED:

1. **Step 1 (commands):** [QUICKSTART.md](../QUICKSTART.md)
2. **Step 2 (guided onboarding):** [GETTING_STARTED.md](GETTING_STARTED.md)
3. **Step 3 (full docs index):** this page (pick a category below)

## ⭐ High-Value Categories

- **API Integration:** [Quick Starts](quickstarts/) • [Concepts](concepts.md) • [CLI Reference](cli-reference.md)
- **Protocol & Policy:** [Token Format](protocol/token-format.md) • [Verification Rules](protocol/verification-rules.md) • [Policies Overview](policies/overview.md)
- **Deployment:** [Docker Compose](deployment/docker-compose.md) • [Linux VM](deployment/vm.md) • [Kubernetes](deployment/kubernetes.md)
- **Operations:** [On-Call Runbook](operations/on-call-runbook.md) • [Troubleshooting](TROUBLESHOOTING.md)
- **Security & Compliance:** [Threat Model](security/threat-model.md) • [Key Management](security/key-management.md) • [SOC2 Compliance](security/soc2-compliance.md)

---

## 📖 Documentation by Category

### 🧭 Repository Docs
| Document | Description |
|----------|-------------|
| [README](../README.md) | Project overview, feature summary, and entry points |
| [Architecture](../ARCHITECTURE.md) | System design and package-level data flow |
| [Security Policy](../SECURITY.md) | Vulnerability reporting process and support policy |

### 🎯 Getting Started
| Document | Description |
|----------|-------------|
| [Overview](overview.md) | High-level introduction to OathMesh and core problem statement |
| [Concepts](concepts.md) | Core terminology: Oath Tokens, Issuer, Receiver, Caller, Subject URIs, Claims |
| [CLI Reference](cli-reference.md) | Terminal commands for minting, revoking, and managing tokens |

### 🏛️ Core Concepts
| Document | Description |
|----------|-------------|
| [Concepts](concepts.md) | Architecture building blocks and terminology (Issuer, Receiver, Caller, Subject) |
| [Policies Overview](policies/overview.md) | How policy engines work and configuration philosophy |
| [Policy Examples](policies/examples.md) | Real-world policy code samples in Pkl and JSON formats |
| [Policy Migration](policies/migration.md) | Migrating from API keys or static secrets to policies |

### 🔐 Protocol Reference
| Document | Description |
|----------|-------------|
| [Token Format](protocol/token-format.md) | JWS structure, headers, and MIME types |
| [Claim Reference](protocol/claim-reference.md) | All JWT claims (sub, aud, act, jti, iat, exp, etc.) explained |
| [Verification Rules](protocol/verification-rules.md) | The 14-step verification pipeline and rejection criteria |
| [Error Taxonomy](protocol/error-taxonomy.md) | Standardized error codes and meanings |
| [Audit Events](protocol/audit-events.md) | NDJSON audit log format and event types |

### 🚀 Quick Starts (5-10 mins each)
| Document | Description |
|----------|-------------|
| [Canonical Quick Start](../QUICKSTART.md) | Clone → build → run → mint → protected API call using current compose ports |
| [Protect a Go chi API](quickstarts/protect-chi-api.md) | Add OathMesh middleware to a Go web service |
| [Protect an Express API](quickstarts/protect-express-api.md) | Express.js integration with full example |
| [Protect a Next.js API](quickstarts/protect-nextjs-api.md) | Next.js App Router and Pages Router support |
| [Protect a FastAPI Service](quickstarts/protect-fastapi.md) | Python FastAPI integration |
| [GitHub Actions → Internal API](quickstarts/github-actions-to-internal-api.md) | Mint tokens in CI/CD workflows |

### 🎓 Tutorials (15-30 mins each)
| Document | Description |
|----------|-------------|
| [Getting Started](tutorials/getting-started.md) | End-to-end guide: issuer setup, token minting, receiver validation |
| [gRPC Integration](tutorials/grpc-integration.md) | Using OathMesh with gRPC services |
| [GraphQL Integration](tutorials/graphql-integration.md) | GraphQL API security with OathMesh (Node.js and Python) |
| [CI/CD Machine Identity](tutorials/ci-cd-machine-identity.md) | GitHub Actions, GitLab CI, and other CI systems |

### 🔧 Deployment & Infrastructure
| Document | Description |
|----------|-------------|
| [Docker Compose](deployment/docker-compose.md) | Multi-container local and staging deployments |
| [Linux VM (systemd)](deployment/vm.md) | Production systemd service setup for physical or VM infrastructure |
| [Kubernetes](deployment/kubernetes.md) | StatefulSet manifests, networking, and scaling |
| [TLS Configuration](deployment/tls.md) | Issuer TLS setup, certificate management, and upstream proxy config |
| [Key Rotation](deployment/key-rotation.md) | Rotating Ed25519 keypairs with zero downtime |

### 🛠️ Operations & Maintenance
| Document | Description |
|----------|-------------|
| [On-Call Runbook](operations/on-call-runbook.md) | Incident response, common issues, and resolution steps |
| [Production Checklist](operations/production-checklist.md) | Pre-deployment verification and hardening steps |
| [Alerting Rules](operations/alerting-rules.md) | Prometheus rules, alerts, and thresholds for monitoring |
| [Grafana Dashboards](operations/grafana-dashboards.md) | Pre-built dashboards for metrics and observability |
| [Best Practices](operations/best-practices.md) | Operational guidance for reliability and security |
| [Incident Response](operations/incident-response.md) | Breach procedures, token revocation, and remediation |
| [Key Rotation](operations/key-rotation.md) | Rotating keys during production uptime |
| [Troubleshooting](TROUBLESHOOTING.md) | Cross-cutting runbook for token, issuer, and receiver failures |

### 🆘 Troubleshooting
| Document | Description |
|----------|-------------|
| [Troubleshooting](TROUBLESHOOTING.md) | End-to-end triage flow and common fixes |
| [On-Call Runbook](operations/on-call-runbook.md) | Incident procedures and escalation steps |

### 🔒 Security & Compliance
| Document | Description |
|----------|-------------|
| [Threat Model](security/threat-model.md) | Threat categories, mitigations, and out-of-scope threats |
| [Key Management](security/key-management.md) | Private key storage, rotation, and hardware security modules (HSM) |
| [Replay Defense](security/replay-defense.md) | How `jti` and replay caches prevent token reuse |
| [Logging Guidance](security/logging-guidance.md) | Sensitive data handling, PII masking, and retention policies |
| [SOC2 Compliance](security/soc2-compliance.md) | Mapping to SOC2 Trust Service Criteria (CC, A, C, S) |
| [GDPR Data Retention](compliance/gdpr-data-retention.md) | Personal data handling and audit log retention requirements |
| [Privacy Controls](compliance/privacy-operational-controls.md) | Privacy by design and operational safeguards |

### 🧩 Architecture & Decisions
| Document | Description |
|----------|-------------|
| [ADR-001: Module Structure](decisions/ADR-001-module-structure.md) | Why OathMesh is organized as it is |
| [ADR-002: Tech Stack](decisions/ADR-002-tech-stack.md) | Go, Redis, Pkl, Ed25519 — and why |
| [ADR-003: Crypto Selection](decisions/ADR-003-crypto-selection.md) | Why Ed25519 over RSA, ECDSA, etc. |

### 🎯 Examples & Samples
| Document | Description |
|----------|-------------|
| [Policy Examples](policies/examples.md) | Pkl and JSON policy files for common scenarios |

---

## 🗺️ Navigation Tips

- **New to OathMesh?** Start with [QUICKSTART.md](../QUICKSTART.md) → [GETTING_STARTED.md](GETTING_STARTED.md) → this index, then pick [Overview](overview.md) or a quickstart
- **Want to deploy?** Jump to [Deployment](deployment/) based on your infrastructure (Docker, VM, or K8s)
- **Need to operate it?** See [Operations](operations/) for monitoring, alerting, and incident response
- **Securing your API?** Pick your framework: [Go](quickstarts/protect-chi-api.md), [Express](quickstarts/protect-express-api.md), [Next.js](quickstarts/protect-nextjs-api.md), [FastAPI](quickstarts/protect-fastapi.md)
- **Protocol deep-dive?** Read [Protocol Reference](protocol/) for token structure, claims, and verification rules
- **Policy questions?** See [Policies](policies/) for configuration examples and migration guides

---

## 📋 Document Status

All documents are production-ready unless marked with:
- 🚧 **In Progress** — May change before next release
- ⚠️ **Requires Security Audit** — Use carefully in production until audit completes

---

**Last updated:** See [CHANGELOG.md](../CHANGELOG.md) for recent changes.
