# OathMesh Kubernetes Manifests

> Production-oriented baseline manifests for Kubernetes 1.24+.

## Included Manifests

- `namespace.yaml`
- `issuer-configmap.yaml`
- `issuer-secret.example.yaml`
- `redis-statefulset.yaml`
- `issuer-service.yaml`
- `issuer-statefulset.yaml`
- `receiver-deployment.yaml`
- `issuer-ingress.yaml`
- `network-policy.yaml`
- `pod-disruption-budget.yaml`

## Prerequisites

- Kubernetes cluster (1.24+)
- `kubectl` configured to the target cluster
- Ingress controller (example uses class `nginx`)
- TLS certificate secret for issuer host

## 1) Configure values

Before applying manifests, update these placeholders:

1. `issuer-configmap.yaml`
   - `OATHMESH_ISSUER=https://issuer.example.com`
2. `issuer-ingress.yaml`
   - `issuer.example.com`
   - `oathmesh-issuer-tls`
3. `receiver-deployment.yaml`
   - `ghcr.io/acme/oathmesh-receiver:latest`
   - `OATHMESH_AUDIENCE`
   - `OATHMESH_TRUSTED_ISSUERS`
4. `redis-statefulset.yaml`
   - `storageClassName: standard` if your cluster uses another class

## 2) Create real issuer secret (do not use the example as-is)

`issuer-secret.example.yaml` is intentionally non-functional.

If the namespace does not exist yet:

```bash
kubectl apply -f namespace.yaml
```

Create a real secret from your secure values:

```bash
kubectl -n oathmesh create secret generic oathmesh-issuer-secrets \
  --from-literal=OATHMESH_MINT_SECRET='<strong-random-secret>' \
  --from-file=OATHMESH_PRIVATE_KEY=./private.pem \
  --dry-run=client -o yaml | kubectl apply -f -
```

## 3) Deploy

```bash
kubectl apply -f namespace.yaml
kubectl apply -f issuer-configmap.yaml
kubectl apply -f redis-statefulset.yaml
kubectl apply -f issuer-service.yaml
kubectl apply -f issuer-statefulset.yaml
kubectl apply -f receiver-deployment.yaml
kubectl apply -f issuer-ingress.yaml
kubectl apply -f network-policy.yaml
kubectl apply -f pod-disruption-budget.yaml
```

## 4) Verify

```bash
kubectl -n oathmesh get pods
kubectl -n oathmesh get svc
kubectl -n oathmesh get ingress
kubectl -n oathmesh get pdb
```

Check issuer health and JWKS:

```bash
kubectl -n oathmesh port-forward svc/oathmesh-issuer 4000:4000
curl http://127.0.0.1:4000/healthz
curl http://127.0.0.1:4000/.well-known/jwks.json
```

Check Redis:

```bash
kubectl -n oathmesh exec -it statefulset/oathmesh-redis -- redis-cli ping
```

## 5) Cleanup

```bash
kubectl delete -f pod-disruption-budget.yaml
kubectl delete -f network-policy.yaml
kubectl delete -f issuer-ingress.yaml
kubectl delete -f receiver-deployment.yaml
kubectl delete -f issuer-statefulset.yaml
kubectl delete -f issuer-service.yaml
kubectl delete -f redis-statefulset.yaml
kubectl delete -f issuer-configmap.yaml
kubectl delete namespace oathmesh
```

## Notes

- Keep issuer secret management in your private GitOps or external secrets system.
- Pin image digests in production.
- Tighten `network-policy.yaml` to your ingress controller namespace/labels.

