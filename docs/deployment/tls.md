← [Back to Index](../INDEX.md)

# TLS Deployment Guide

> TLS for OathMesh issuer endpoints, aligned with current runtime/env configuration.

## Why This Matters

Receivers fetch signing keys from `${OATHMESH_ISSUER}/.well-known/jwks.json`.  
If that URL is intercepted, an attacker can swap keys and mint tokens that appear valid.

Production rule: use `https://` issuer URLs and terminate TLS at a trusted edge (Ingress/LB/proxy).

## OathMesh Runtime Expectations

- Issuer process serves HTTP on `OATHMESH_PORT` (default `4000`).
- `OATHMESH_ISSUER` is the canonical external URL used in tokens/discovery/JWKS metadata.
- In non-development environments (`OATHMESH_ENV != development`), startup validation requires `OATHMESH_ISSUER` to use `https://`.
- Receiver trust values must match the same issuer URL:
  - `OATHMESH_TRUSTED_ISSUERS` (receiver SDK apps)
  - `OATHMESH_GATEWAY_ISSUERS` (gateway mode)

Example (Kubernetes ConfigMap pattern in this repo):

```yaml
data:
  OATHMESH_ISSUER: https://issuer.example.com
  OATHMESH_ENV: production
  OATHMESH_PORT: "4000"
```

## Local/Dev: Self-Signed TLS

For local testing that mirrors production trust behavior, keep issuer on HTTP internally and put TLS in front.

### 1) Generate a local cert

```bash
openssl req -x509 -nodes -newkey rsa:2048 \
  -keyout issuer-local.key \
  -out issuer-local.crt \
  -days 365 \
  -subj "/CN=issuer.localhost"
```

### 2) Run issuer normally

```bash
OATHMESH_ENV=development \
OATHMESH_ISSUER=https://issuer.localhost \
OATHMESH_PRIVATE_KEY_FILE=./private.pem \
OATHMESH_MINT_SECRET=development_secret_do_not_use_in_prod \
go run ./cmd/oathmesh serve --port 4000
```

### 3) Front it with local TLS termination (nginx example)

```nginx
server {
  listen 443 ssl;
  server_name issuer.localhost;

  ssl_certificate     /etc/nginx/certs/issuer-local.crt;
  ssl_certificate_key /etc/nginx/certs/issuer-local.key;
  ssl_protocols TLSv1.2 TLSv1.3;

  location / {
    proxy_pass http://host.docker.internal:4000;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto https;
  }
}
```

### 4) Point verifiers at HTTPS issuer

```bash
OATHMESH_TRUSTED_ISSUERS=https://issuer.localhost
# or gateway mode:
OATHMESH_GATEWAY_ISSUERS=https://issuer.localhost
```

## Production Certificate Management Options

### 1) ACME automation (cert-manager / Let’s Encrypt)

Best for internet-reachable DNS names. Common in Kubernetes with Ingress TLS secret rotation.

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: oathmesh-issuer-tls
  namespace: oathmesh
spec:
  secretName: oathmesh-issuer-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - issuer.example.com
```

### 2) Ingress / Load Balancer termination

Recommended default for this repo’s deployment model:

```text
Client -> TLS edge (Ingress/LB) -> HTTP :4000 issuer pod/service
```

Use `examples/kubernetes/issuer-ingress.yaml` pattern:
- `force-ssl-redirect: "true"`
- `spec.tls[].secretName: oathmesh-issuer-tls`

### 3) Internal PKI

For private networks and enterprise CAs:
- Issue certs from your internal CA (Vault PKI, AD CS, step-ca, etc.).
- Install CA trust in all calling workloads/verifiers.
- Keep `OATHMESH_ISSUER` as the exact HTTPS URL served by that cert.

## Issuer HTTPS Configuration Specifics

1. Set canonical issuer URL:

```bash
OATHMESH_ISSUER=https://issuer.example.com
```

2. Set production mode:

```bash
OATHMESH_ENV=production
```

3. Keep internal listen port as plain HTTP (default):

```bash
OATHMESH_PORT=4000
```

4. Keep receiver issuer trust in sync:

```bash
OATHMESH_TRUSTED_ISSUERS=https://issuer.example.com
# or
OATHMESH_GATEWAY_ISSUERS=https://issuer.example.com
```

5. Verify JWKS is reachable over TLS:

```bash
curl -fsS https://issuer.example.com/.well-known/jwks.json
```

## mTLS Guidance (Internal Service-to-Service)

OathMesh token verification and mTLS solve different layers:
- OathMesh: call identity + scoped authorization context.
- mTLS: transport peer authentication + channel encryption.

Where to apply mTLS:
- Between internal services (east-west traffic).
- Between ingress gateway and upstream services if required by policy.

How to apply with current runtime:
- Keep OathMesh issuer app config unchanged (`OATHMESH_PORT`, `OATHMESH_ISSUER`, etc.).
- Enforce mTLS at your platform layer (service mesh, ingress controller, or sidecar proxy).
- Continue using HTTPS issuer URL for JWKS trust; mTLS is additive, not a replacement.

## Verification Checklist

```bash
# External TLS + cert chain
curl -I https://issuer.example.com/.well-known/jwks.json
openssl s_client -connect issuer.example.com:443 -servername issuer.example.com < /dev/null

# Issuer URL values in cluster
kubectl -n oathmesh get configmap oathmesh-issuer-config -o yaml

# Ensure issuer app still healthy internally
kubectl -n oathmesh port-forward svc/oathmesh-issuer 4000:4000
curl -f http://127.0.0.1:4000/healthz
```

## Troubleshooting

### `OATHMESH_ISSUER must use HTTPS in non-development environments`
- Cause: `OATHMESH_ENV=production` with `http://` issuer URL.
- Fix: set `OATHMESH_ISSUER=https://...` and ensure TLS is actually served there.

### Receiver errors like `issuer_untrusted`
- Cause: issuer mismatch (`iss` claim vs configured trusted issuers).
- Fix: align `OATHMESH_TRUSTED_ISSUERS` / `OATHMESH_GATEWAY_ISSUERS` exactly with `OATHMESH_ISSUER`.

### TLS handshake/certificate errors from clients
- Cause: missing CA trust, wrong SAN/CN, expired cert, or host mismatch.
- Fix: issue cert for the exact hostname in `OATHMESH_ISSUER`; install the correct CA chain.

### JWKS reachable on HTTP but not HTTPS
- Cause: edge TLS route/cert secret misconfigured.
- Fix: verify ingress/LB listener on 443, certificate secret, and host rule mapping to issuer service.

## Related

- [Kubernetes Deployment Guide](kubernetes.md)
- [Docker Compose Deployment](docker-compose.md)
- [Issuer Configuration Reference](../config/issuer-config.md)
- [Production Go-Live Checklist](../operations/production-checklist.md)
