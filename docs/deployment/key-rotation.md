# Key Rotation Guide

> Procedures for rotating signing keys in OathMesh.

## Overview

OathMesh uses Ed25519 keys for token signing. Key rotation is critical for:
1. Limiting exposure if a key is compromised
2. Compliance requirements
3. Regular security hygiene

## Key Format

OathMesh keys use the format: `issuer-key-YYYY-MM-{4-char-random-hex}`

Example: `issuer-key-2025-04-a3f2`

## Rotation Procedures

### Manual Rotation

1. **Generate new key pair:**
   ```bash
   openssl genpkey -algorithm Ed25519 -out new-key.pem
   ```

2. **Add new key to issuer:**
   ```bash
   ./bin/oathmesh key add --key new-key.pem --kid issuer-key-2025-04-a3f2
   ```

3. **Verify JWKS includes new key:**
   ```bash
   curl https://issuer.oathmesh.internal/.well-known/jwks.json
   ```

4. **Update client configurations** to point to new issuer URL if kid changed

5. **Remove old key** after grace period (typically 24-72 hours):
   ```bash
   ./bin/oathmesh key revoke --kid issuer-key-2024-03-b1c2
   ```

### Automated Rotation

For automated rotation with cert-manager:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: oathmesh-signing-key
spec:
  secretName: oathmesh-signing-key
  issuerRef:
    name: internal-ca
    kind: Issuer
  commonName: oathmesh-signing
  dnsNames:
  - issuer.oathmesh.internal
```

### Kubernetes Secrets Rotation

```bash
# Create new key secret
kubectl create secret generic oathmesh-signing-key-new \
  --from-file=private-key=new-key.pem \
  -n oathmesh --dry-run=client -o yaml | kubectl apply -f -

# Rolling restart to pick up new key
kubectl rollout restart deployment/oathmesh-issuer -n oathmesh

# After grace period, remove old secret
kubectl delete secret oathmesh-signing-key-old -n oathmesh
```

## Grace Period

After adding a new key, maintain the old key for a grace period:
- **Minimum**: 24 hours
- **Recommended**: 72 hours
- **High-security**: 7 days

This allows clients time to fetch the new JWKS without token validation failures.

## Monitoring

Monitor for key usage:
```bash
# Check active keys
./bin/oathmesh key list

# Audit log for key operations
kubectl logs -n oathmesh -l app=oathmesh-issuer | grep -i key
```

## Emergency Revocation

If a key is compromised:

1. **Immediately revoke the compromised key:**
   ```bash
   ./bin/oathmesh key revoke --kid <compromised-kid>
   ```

2. **Generate emergency replacement:**
   ```bash
   openssl genpkey -algorithm Ed25519 -out emergency-key.pem
   ./bin/oathmesh key add --key emergency-key.pem
   ```

3. **Push updated JWKS:**
   ```bash
   # Force JWKS cache invalidation across all gateways
   kubectl rollout restart deployment/oathmesh-gateway -n oathmesh
   ```

4. **Investigate the compromise** before resuming service

## Rotation Checklist

- [ ] Generate new key pair
- [ ] Add new key to issuer with unique kid
- [ ] Verify new key appears in JWKS
- [ ] Update client configurations if needed
- [ ] Monitor for successful validation with new key
- [ ] Wait grace period (24-72 hours)
- [ ] Revoke old key
- [ ] Archive old key securely or destroy

## Automated Rotation Schedule

Set up cron for regular rotation:

```bash
# crontab entry - rotate monthly
0 0 1 * * /path/to/oathmesh key rotate --schedule monthly
```

Or use Kubernetes Job:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: oathmesh-key-rotation
spec:
  schedule: "0 0 1 * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: rotate
            image: oathmesh/oathmesh:latest
            command: ["/bin/oathmesh", "key", "rotate"]
          restartPolicy: OnFailure