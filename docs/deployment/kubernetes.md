← [Back to Index](../INDEX.md)

# Kubernetes Deployment Guide

> Production-oriented Kubernetes deployment for OathMesh issuer + Redis replay cache and a protected receiver.

## Overview

Use the manifests in [`examples/kubernetes`](../../examples/kubernetes) as a hardened baseline:

- Kubernetes 1.24+ compatible APIs
- Non-root containers and restricted security context
- Readiness/liveness probes and resource requests/limits
- NetworkPolicy and PodDisruptionBudget
- Explicit OathMesh environment variable wiring

## Reference Topology

```
Internet
   │
   ▼
Ingress (TLS)
   │
   ▼
oathmesh-issuer (StatefulSet) ─────► Redis (StatefulSet)
        │
        └──── JWKS consumed by receiver service(s)
```

## What to Customize Before Deploying

1. **Images**
   - `oathmesh/oathmesh:latest` for issuer (or your pinned digest)
   - `ghcr.io/acme/oathmesh-receiver:latest` placeholder in receiver deployment
2. **DNS/TLS**
   - Set Ingress host in `issuer-ingress.yaml`
   - Use your real TLS secret name
3. **Secrets**
   - Copy `issuer-secret.example.yaml` to your private ops repo and fill real values
4. **Storage class**
   - Update Redis PVC `storageClassName` if needed

## OathMesh Environment Wiring

Issuer settings are split so non-sensitive values stay in ConfigMap and secrets stay in Secret:

- `issuer-configmap.yaml`
  - `OATHMESH_ISSUER` (must be `https://...` in production)
  - `OATHMESH_ENV=production`
  - TTL/rate-limit defaults
  - `REDIS_URL=redis://oathmesh-redis:6379/0`
- `issuer-secret.example.yaml`
  - `OATHMESH_PRIVATE_KEY` (PEM)
  - `OATHMESH_MINT_SECRET`

Receiver wiring (example):

- `OATHMESH_AUDIENCE=https://inventory.internal`
- `OATHMESH_TRUSTED_ISSUERS=https://issuer.example.com`

## Deployment Order

```bash
kubectl apply -f examples/kubernetes/namespace.yaml
kubectl apply -f examples/kubernetes/issuer-configmap.yaml
kubectl -n oathmesh create secret generic oathmesh-issuer-secrets \
  --from-literal=OATHMESH_MINT_SECRET='<strong-random-secret>' \
  --from-file=OATHMESH_PRIVATE_KEY=./private.pem \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f examples/kubernetes/redis-statefulset.yaml
kubectl apply -f examples/kubernetes/issuer-service.yaml
kubectl apply -f examples/kubernetes/issuer-statefulset.yaml
kubectl apply -f examples/kubernetes/receiver-deployment.yaml
kubectl apply -f examples/kubernetes/issuer-ingress.yaml
kubectl apply -f examples/kubernetes/network-policy.yaml
kubectl apply -f examples/kubernetes/pod-disruption-budget.yaml
```

## Verification

```bash
kubectl -n oathmesh get pods
kubectl -n oathmesh get svc
kubectl -n oathmesh get ingress
kubectl -n oathmesh logs statefulset/oathmesh-issuer --tail=100
```

Quick health checks:

```bash
kubectl -n oathmesh port-forward svc/oathmesh-issuer 4000:4000
curl http://127.0.0.1:4000/healthz
curl http://127.0.0.1:4000/.well-known/jwks.json
```

## Security Notes

- Never commit real secrets or private keys.
- Keep issuer ingress locked down to trusted callers (WAF/IP allowlist, authn at edge, or private ingress).
- Use Redis persistence + backups for revocation continuity.
- Pin image digests for production change control.
- Keep `OATHMESH_ISSUER` and receiver `OATHMESH_TRUSTED_ISSUERS` aligned.

## Related

- [TLS Configuration Guide](tls.md)
- [Issuer Configuration Reference](../config/issuer-config.md)
- [Examples: Kubernetes manifests](../../examples/kubernetes/README.md)
