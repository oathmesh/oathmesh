# 📚 OathMesh Documentation Index

Welcome to OathMesh! Use this index to navigate the complete documentation. Quick links are below for your role.

---

## 🚀 Quick Links by Role

**👤 First-Time User?**
- [OathMesh Overview](overview.md) — What it is and why you need it
- [Getting Started Tutorial](tutorials/getting-started.md) — Build your first issuer + receiver
- [Docker Quick Start](quickstarts/local-demo-docker-compose.md) — Get running in 5 minutes

**⚙️ Developer / API Owner?**
- [Core Concepts](concepts.md) — Terminology and architecture
- [Protect a Go API](quickstarts/protect-chi-api.md) | [Express](quickstarts/protect-express-api.md) | [Next.js](quickstarts/protect-nextjs-api.md) | [FastAPI](quickstarts/protect-fastapi.md)
- [Token Format Reference](protocol/token-format.md) — What's inside a token
- [Policy Overview](policies/overview.md) — How to configure access rules
- [CLI Reference](cli-reference.md) — Terminal commands
- [Troubleshooting Guide](TROUBLESHOOTING.md) — Fast diagnosis for common failures

**🏗️ DevOps / Platform Engineer?**
- [Deployment Options](deployment/docker-compose.md) — Docker, VM, or Kubernetes
- [Linux VM Setup](deployment/vm.md) — systemd-based production deployment
- [Kubernetes Guide](deployment/kubernetes.md) — K8s manifests and scaling
- [Operations Runbook](operations/on-call-runbook.md) — Troubleshooting and incidents
- [Key Rotation](deployment/key-rotation.md) — Rolling keys safely

**🔒 Security Engineer / Auditor?**
- [Threat Model](security/threat-model.md) — What OathMesh protects against
- [Security Guide](security/key-management.md) — Key management and secrets handling
- [Verification Rules](protocol/verification-rules.md) — The 14-step verification pipeline
- [Replay Defense](security/replay-defense.md) — How replay attacks are prevented
- [SOC2 Compliance Matrix](security/soc2-compliance.md) — Compliance mappings
- [Logging Guidance](security/logging-guidance.md) — What to log and why

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
| [Local Demo (Docker Compose)](quickstarts/local-demo-docker-compose.md) | Spin up issuer + receiver locally in <2 minutes |
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

- **New to OathMesh?** Start with [Overview](overview.md) → [Concepts](concepts.md) → [Getting Started Tutorial](tutorials/getting-started.md)
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
