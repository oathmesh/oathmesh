# 🚀 Getting Started with OathMesh

<p align="center">
  <img src="../assets/logo.png" width="120" alt="OathMesh Logo">
</p>

<p align="center">
  <b>5–10 minute walkthrough from zero to a working, protected API call.</b>
</p>

---

> 👋 **New to OathMesh?** You're in the right place. This guide will get you up and running with a hands-on example in minutes.
>
> After this, explore the [Concepts](concepts.md) doc for deeper understanding, or jump to [Architecture](../ARCHITECTURE.md) for system design.

## 🚦 Start Here (Canonical Flow)

Follow this canonical Start Here flow used by the README and docs index:

1. **Step 1 (commands):** [QUICKSTART.md](../QUICKSTART.md)
2. **Step 2 (guided onboarding):** this page
3. **Step 3 (full docs index):** [INDEX.md](INDEX.md)

Before production rollout, complete the [operations checklist](operations/production-checklist.md) and review [SECURITY.md](../SECURITY.md).

## Prerequisites

Before you start, make sure you have:

- **Docker & Docker Compose** — for running the issuer and demo
- **Go toolchain that satisfies `go.mod`** (currently `go 1.26.2`) — for building the CLI from source
- **curl** and **jq** — for testing API calls
- **openssl** — for generating Ed25519 keys

## After Step 1, Choose Your Path

Choose your path:

### Option A: Run the Local Demo (5 minutes) 🐳

**Best for:** Understanding OathMesh by seeing it in action locally.

You'll start an issuer + protected demo API, mint a token, call the API, and validate replay/expiry behavior.

Use the canonical commands from Step 1 (`QUICKSTART.md`) for the current compose mapping:
- issuer: `http://localhost:4000`
- protected chi API: `http://localhost:8081/inventory`

👉 **[Run the canonical local demo path →](../QUICKSTART.md)**

---

### Option B: Protect Your Existing API (10–15 minutes) 🛡️

**Best for:** Adding OathMesh token verification to your own service.

Choose your framework:

| Framework | Time | Setup | Guide |
|-----------|------|-------|-------|
| **Express.js** | ~5 min | Node 18+ | [Protect Express API →](quickstarts/protect-express-api.md) |
| **Next.js** | ~5 min | Next.js 13+ | [Protect Next.js API →](quickstarts/protect-nextjs-api.md) |
| **FastAPI / Flask** | ~5 min | Python 3.9+ | [Protect FastAPI →](quickstarts/protect-fastapi.md) |
| **Go chi** | ~5 min | Go toolchain matching `go.mod` | [Protect Go chi API →](quickstarts/protect-chi-api.md) |
| **GitHub Actions** | ~15 min | GH OIDC | [GitHub Actions CI/CD →](quickstarts/github-actions-to-internal-api.md) |

Each guide shows you:
1. Installing the SDK
2. Configuring the middleware
3. Testing with a real token


## SDK Capability Matrix

| Capability | Go SDK | Node SDK | Python SDK |
|---|---|---|---|
| HTTP middleware integration | ✅ chi + `net/http` | ✅ Express | ✅ FastAPI/Flask/Django patterns |
| Next.js support | ❌ | ✅ App/Pages/Edge | ❌ |
| gRPC server interceptors | ✅ unary + stream | ❌ | ❌ |
| Core verifier usable in any runtime | ✅ | ✅ | ✅ |
| Lifecycle hooks (`onVerified`/`onDenied`) | ✅ (verifier callbacks) | ✅ | ✅ |
| Auto-mint client helper | ❌ | ✅ `OathMeshClient` | ✅ `OathMeshClient` |


---

### Option C: Understand How It Works First (10 minutes) 📖

**Best for:** Learning the concepts before diving into code.

Start here for a solid foundation:

1. **[🔗 Core Concepts](concepts.md)** — Oath Tokens, Issuers, Callers, Receivers, and the VerifiedCallerContext
2. **[🏗️ Architecture](../ARCHITECTURE.md)** — System design, package structure, and data flow
3. **[🔐 Verification Pipeline](../docs/protocol/verification-rules.md)** — The 14-step verification process
4. **[🛡️ Threat Model](../docs/security/threat-model.md)** — Security guarantees and what OathMesh protects against

Then move on to **Option A** or **Option B** above.

---

## Next Steps

### For Developers (Building with OathMesh)

Once you've completed your chosen path above:

- **Deep dive into SDKs:**
  - [Go SDK Reference](../sdk/go/middleware/README.md)
  - [Node.js SDK Reference](../sdk/node/README.md)
  - [Python SDK Reference](../sdk/python/README.md)

- **Advanced topics:**
  - [Policy Configuration Guide](../docs/config/pkl-policy-guide.md) — Write custom authorization rules
  - [Token Format & Claims](../docs/protocol/token-format.md) — Full token structure
  - [Error Taxonomy](../docs/protocol/error-taxonomy.md) — All error codes and how to handle them
  - [CLI Reference](../docs/cli-reference.md) — Token minting and verification commands

---

### For Operators (Deploying OathMesh)

Once your application is protected, deploy the issuer:

- **[Linux VM Deployment (systemd)](../docs/deployment/vm.md)** — Single-node production setup
- **[Docker Compose Deployment](../docs/deployment/docker-compose.md)** — Multi-service orchestration
- **[Kubernetes Deployment](../docs/deployment/kubernetes.md)** — Cloud-native scaling
- **[TLS Configuration](../docs/deployment/tls.md)** — Securing the issuer in transit

### Deployment Option Comparison

| Option | Best For | HA Pattern | Ops Overhead | Typical Time-to-First-Deploy |
|---|---|---|---|---|
| **Docker Compose** | Local dev, PoC, small internal environments | Single host restart policy | Low | 10-20 min |
| **Kubernetes** | Multi-team production, autoscaling, policy-heavy environments | Multi-replica pods + managed Redis | High | 1-2 hours |
| **Linux VM (systemd)** | Traditional ops teams, predictable fixed workloads | Active/passive + external LB | Medium | 30-60 min |

- **Production readiness:**
  - [Production Checklist](../docs/operations/production-checklist.md) — Pre-launch tasks
  - [Key Rotation Guide](../docs/operations/key-rotation.md) — Signing key management
  - [Alerting & Monitoring](../docs/operations/grafana-dashboards.md) — Observability setup
  - [Incident Response Runbook](../docs/operations/incident-response.md) — On-call playbook

---

## Troubleshooting

For fast triage and production-friendly debugging, use the dedicated troubleshooting guides:

- [Global Troubleshooting Guide](TROUBLESHOOTING.md)
- [chi example troubleshooting](../examples/chi-api/TROUBLESHOOTING.md)
- [Express example troubleshooting](../examples/express-api/TROUBLESHOOTING.md)
- [FastAPI example troubleshooting](../examples/fastapi-api/TROUBLESHOOTING.md)
- [Next.js example troubleshooting](../examples/nextjs-api/TROUBLESHOOTING.md)
- [gRPC server troubleshooting](../examples/grpc-server/TROUBLESHOOTING.md)

Need community support? See [SUPPORT.md](../SUPPORT.md).

---

## Core Doctrine

> **"OathMesh authenticates the caller. The receiver authorizes the request."**

OathMesh tells you *who* is calling, *what* they want to do, and *where* the call came from. Your application decides whether to allow it. This separation of concerns means:

- ✅ You keep full control over authorization policy
- ✅ OathMesh can't accidentally grant access to something it shouldn't
- ✅ Your business logic remains in your code, not in a third-party system

---

## What's Next After Getting Started?

| Goal | Go to |
|------|-------|
| Learn what OathMesh is and when to use it | [Overview](overview.md) |
| Understand core concepts and terminology | [Concepts](concepts.md) |
| See system design and architecture | [ARCHITECTURE.md](../ARCHITECTURE.md) |
| Configure policies for your use case | [Policy Guide](../docs/config/pkl-policy-guide.md) |
| Deploy to production | [Deployment Guides](../docs/deployment/) |
| Understand compliance & security | [Security Docs](../docs/security/) |

---

## Quick Links

- 🔗 [GitHub Repository](https://github.com/oathmesh/oathmesh)
- 📖 [Full Documentation Index](../docs/)
- 🐛 [Report an Issue](https://github.com/oathmesh/oathmesh/issues)
- 💬 [Community & Support](../SUPPORT.md)
- 📋 [Contributing Guide](../CONTRIBUTING.md)
