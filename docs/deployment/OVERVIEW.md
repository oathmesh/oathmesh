---
title: Deployment Overview
description: Choose the right OathMesh deployment model for local development, VM environments, Kubernetes, or edge TLS layers.
tags: [deployment, docker, kubernetes, vm, tls, operations]
---

← [Back to Index](../INDEX.md)

# Deployment Decision Tree

<p align="center">
  <b>Quick guide to choosing the right OathMesh deployment strategy.</b>
</p>

---

## Deployment Decision Flow

```
START: "Where will OathMesh run?"
│
├─────────────────────────────────────────────────┐
│                                                 │
▼                                                 ▼
Local Dev?                              Production / Self-Hosted?
│                                                 │
│                                                 │
├─→ Docker Compose                      ├─→ Kubernetes
│   (Single-machine, dev)                │   (Multi-cluster, enterprise)
│   │                                    │
│   ├─ Issuer container                 ├─ Issuer StatefulSet (1 replica)
│   ├─ Redis in-memory                  ├─ Redis Deployment (HA optional)
│   ├─ Optional test receiver            ├─ NetworkPolicy isolation
│   ├─ HTTPS via mkcert/local-ca         ├─ Secrets management (sealed/vault)
│   └─ Hot-reload policy files           ├─ Resource quotas/limits
│                                        ├─ Liveness/readiness probes
│       See: docker-compose.md           └─ PodDisruptionBudget
│                                            See: kubernetes.md
│
├─────────────────────────────────────────────────┐
│                                                 │
└─→ Virtual Machines / Bare Metal               └─→ TLS Layer / Reverse Proxy
    (Linear scaling, legacy ops)                   (Caching, WAF, legacy access)
    │                                              │
    ├─ Binary or systemd                          ├─ Ngrok, Cloudflare Tunnel
    ├─ Manual key rotation                        ├─ Application Load Balancer
    ├─ File-based policy persistence              ├─ TLS termination
    ├─ Separate Redis instance                    ├─ Request signing/validation
    └─ SSH access for ops                         ├─ Legacy HTTPS bridge
                                                  └─ Cloud CDN integration
        See: vm.md                                    See: tls.md
```

## Deployment Options Comparison (K8s vs Compose vs VM)

| Criterion | Docker Compose | Kubernetes | Linux VM (systemd) |
|---|---|---|---|
| Best use case | Local dev, staging, small internal deployments | Production at scale, multi-team platforms | Self-hosted production with traditional ops |
| Initial complexity | Low | High | Medium |
| Horizontal scaling | Manual (host-level) | Native (HPA/replicas) | Manual (add VMs + LB) |
| High availability | Limited (single host) | Strong (multi-node/zone capable) | Medium (active/passive or N+1) |
| Secrets management | `.env` + host controls | Kubernetes Secrets / external secret stores | OS secret store / vault integration |
| Operational overhead | Low | High | Medium |
| Time to first deploy | 10-20 min | 1-2 hours | 30-60 min |
| Recommended docs | [docker-compose.md](docker-compose.md) | [kubernetes.md](kubernetes.md) | [vm.md](vm.md) |

> Need edge TLS, WAF, or reverse proxy patterns? Add [tls.md](tls.md) regardless of your base deployment choice.

## Selecting Your Path

**→ Start here:** Is this development, PoC, or production?

- **Local Development:** Use [Docker Compose](docker-compose.md)
- **Self-Hosted Production (Kubernetes available):** Use [Kubernetes](kubernetes.md)
- **Self-Hosted Production (No Kubernetes):** Use [VMs](vm.md)
- **Cloud-Native / Hybrid:** Use [TLS Layer](tls.md) for edge deployment

---

## Common Topologies

### 1. Single-Machine Dev (Docker Compose)
```
┌──────────────────────────┐
│     Laptop / Desktop     │
│                          │
│  ┌────────┐  ┌────────┐  │
│  │ Issuer │  │ Redis  │  │
│  └────────┘  └────────┘  │
│                          │
│  docker-compose up       │
└──────────────────────────┘
```

### 2. Production Kubernetes
```
┌──────────────────────────────────────────┐
│        Kubernetes Cluster (HA)           │
│                                          │
│  ┌─────────────┐      ┌─────────────┐   │
│  │   Issuer    │      │    Redis    │   │
│  │ StatefulSet │      │ Deployment  │   │
│  └─────────────┘      └─────────────┘   │
│                                          │
│  ← Ingress / LoadBalancer                │
└──────────────────────────────────────────┘
```

### 3. VM-Based Production
```
┌─────────────────┬─────────────────┬─────────────────┐
│   Issuer VM 1   │   Issuer VM 2   │   Redis VM      │
│  oathmesh bin   │  oathmesh bin   │  redis-server   │
│  (active)       │  (standby)      │  (primary)      │
└─────────────────┴─────────────────┴─────────────────┘
         │               │                  │
         └───────────────┴──────────────────┘
                   Load Balancer
```

---

## Related Documentation

| Guide | Best For | Effort |
|-------|----------|--------|
| [Docker Compose](docker-compose.md) | Local dev + small deployments | ~15 min |
| [Kubernetes](kubernetes.md) | Enterprise, HA, auto-scaling | ~1 hr |
| [VMs](vm.md) | Self-hosted, legacy ops | ~30 min |
| [TLS Layer](tls.md) | Hybrid, edge, legacy bridge | ~45 min |
| [Production Checklist](../operations/production-checklist.md) | Production readiness validation | ~20 min |
