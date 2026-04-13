# TLS Configuration Guide

> Configuring TLS for OathMesh in production.

## Why TLS is Required

OathMesh tokens are signed with Ed25519, but the **JWKS fetch** that receivers use to get public keys happens over HTTP. If an attacker can intercept this traffic, they can:

1. Perform a man-in-the-middle attack on JWKS fetch
2. Inject their own public key into the response
3. Sign tokens that receivers will trust

**This is a critical threat.** Production deployments MUST use HTTPS for the issuer URL.

## Issuer URL Configuration

Set `OATHMESH_ISSUER` to an HTTPS URL:

```bash
# ✅ Correct - Production
OATHMESH_ISSUER=https://issuer.oathmesh.internal

# ❌ Wrong - Production (exposes JWKS to interception)
OATHMESH_ISSUER=http://issuer.internal
```

## TLS Termination Options

### Option 1: Load Balancer / Cloud LB (Recommended)

Most cloud providers (AWS ALB, GCP Cloud Load Balancer, Azure Load Balancer) provide TLS termination. Configure:

```
Client → [TLS] → Load Balancer → [HTTP] → OathMesh Issuer
```

The issuer still runs HTTP internally, but the public endpoint is TLS-protected.

### Option 2: Nginx Sidecar

If you need end-to-end encryption or the issuer is directly exposed:

```nginx
server {
    listen 443 ssl;
    server_name issuer.oathmesh.internal;

    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;

    # Modern TLS settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://localhost:4000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /.well-known/jwks.json {
        proxy_pass http://localhost:4000;
    }
}
```

### Option 3: Traefik

If using Traefik as your ingress:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: oathmesh-issuer
spec:
  entryPoints:
    - websecure
  tls:
    certResolver: letsencrypt
  routes:
  - match: PathPrefix(`/.well-known/jwks.json`)
    kind: Rule
    services:
    - name: oathmesh-issuer
      port: 4000
```

## Issuer Configuration for TLS

Update your deployment:

```yaml
env:
- name: OATHMESH_ISSUER
  value: "https://issuer.oathmesh.internal"
```

The issuer will automatically serve JWKS at `https://issuer.oathmesh.internal/.well-known/jwks.json`.

## Client-Side TLS (For Issuers Connecting to External Services)

If your issuer needs to connect to external services over TLS (e.g., Redis with TLS, custom webhooks), configure:

```bash
# Custom CA bundle (optional)
OATHMESH_CA_BUNDLE=/path/to/ca-bundle.pem

# Skip TLS verification (ONLY for development)
# OATHMESH_SKIP_TLS_VERIFY=false
```

## Certificate Rotation

### Automatic (Recommended)

Use cert-manager with Let's Encrypt:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: oathmesh-issuer-tls
spec:
  secretName: oathmesh-issuer-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - issuer.oathmesh.internal
```

### Manual

For manual rotation:

1. Get new certificate from your CA
2. Update the Kubernetes secret:
   ```bash
   kubectl create secret tls oathmesh-issuer-tls \
     --cert=new-cert.crt \
     --key=new-cert.key \
     --dry-run=client -o yaml | kubectl apply -f -
   ```
3. Rolling restart of issuer pods:
   ```bash
   kubectl rollout restart deployment/oathmesh-issuer -n oathmesh
   ```

## Minimum TLS Version

Always use **TLS 1.2** or higher. Configure your load balancer/nginx to reject older protocols:

```nginx
ssl_protocols TLSv1.2 TLSv1.3;
```

## Cipher Suites

Use modern cipher suites only:

```nginx
ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
```

## Security Checklist

- [ ] Issuer URL uses https:// in production
- [ ] TLS 1.2 minimum (reject TLS 1.0/1.1)
- [ ] Modern cipher suites configured
- [ ] Certificate auto-rotation configured (cert-manager recommended)
- [ ] Load balancer/ingress enforces HTTPS
- [ ] HTTP to HTTPS redirect configured (if needed)
- [ ] JWT fetch verified over TLS

## Testing TLS Configuration

```bash
# Test JWKS endpoint
curl -I https://issuer.oathmesh.internal/.well-known/jwks.json

# Verify TLS version
openssl s_client -connect issuer.oathmesh.internal:443 -tls1_2

# Check certificate
openssl s_client -showcerts -connect issuer.oathmesh.internal:443 | openssl x509 -noout -text
```