# Incident Response Playbook

**Last Updated:** 2024 | **Status:** Active  
**Owner:** Security & SRE Teams | **Review Cadence:** Quarterly

---

## Overview

This playbook covers response procedures for security and operational incidents affecting OathMesh. All times are target recovery times; actual times depend on incident severity and context.

### Severity Levels

| Level | Definition | Response Time | Recovery Target |
|-------|-----------|---|---|
| 🔴 **Critical** | Security breach or complete service outage | < 5 min | < 30 min |
| 🟠 **High** | Partial service degradation or security concern | < 15 min | < 2 hours |
| 🟡 **Medium** | Non-emergency operational issue | < 1 hour | < 6 hours |
| 🟢 **Low** | Minor issue, documentation or tooling | < 1 day | N/A |

---

## Incident: Private Key Compromised

**Severity:** 🔴 **CRITICAL**  
**Detection Time:** < 5 minutes  
**Recovery Time:** ~10 minutes  
**Owner:** Security Lead + Backend Lead

### Detection Signals

Any of the following indicates potential key compromise:

1. **Unexpected Key ID in tokens**
   - Audit logs show tokens being verified with a different `kid` than expected
   - Indicates an attacker may be signing tokens with a leaked private key

2. **Security alert from external party**
   - Third-party reports tokens they didn't issue
   - Notification of leaked key material

3. **Audit logs show suspicious patterns**
   - Sudden spike in token verification attempts
   - Tokens being verified from unexpected subjects or issuers

### Immediate Response (0-5 minutes)

**Step 1: Confirm the compromise**
```bash
# SSH to issuer host
# Check recent audit logs
tail -f /var/log/oathmesh/audit.log | grep -i "verify"

# Look for:
# - Tokens with unexpected key IDs
# - Authentication from unexpected subjects
# - Unusual verification patterns
```

**Step 2: Stop the issuer service**
```bash
# This prevents further tokens from being signed with the compromised key
docker stop oathmesh-issuer

# Or if using systemd:
systemctl stop oathmesh-issuer
```

**Step 3: Generate new Ed25519 key pair**
```bash
# Generate new keypair (DO NOT use the old one)
openssl genpkey -algorithm ED25519 -out /tmp/new_private.pem

# Extract public key
openssl pkey -in /tmp/new_private.pem -pubout -out /tmp/new_public.pem

# Store public key (will be in JWKS endpoint)
# Keep private key in secure storage (AWS Secrets Manager, HashiCorp Vault, etc.)
```

**Step 4: Update issuer with new key**
```bash
# Option 1: If using environment variable
export OATHMESH_PRIVATE_KEY=$(cat /tmp/new_private.pem)

# Option 2: If using AWS Secrets Manager
aws secretsmanager update-secret --secret-id oathmesh/private-key \
  --secret-string file:///tmp/new_private.pem

# Option 3: If using file-based config
cp /tmp/new_private.pem /etc/oathmesh/private.pem
chmod 400 /etc/oathmesh/private.pem
```

**Step 5: Start issuer with new key**
```bash
docker start oathmesh-issuer
# Or
systemctl start oathmesh-issuer

# Verify startup
docker logs oathmesh-issuer | tail -20
# Look for "issuer started successfully" or similar
```

### Propagation (5-10 minutes)

**Step 6: Force JWKS cache invalidation on all receivers**
```bash
# Option 1: Manually clear cache on each receiver (if you have direct access)
# This depends on your receiver implementation, but typically:
curl -X POST http://localhost:4001/internal/cache/clear

# Option 2: Use monitoring/deployment system
# - If using Kubernetes, restart all receiver pods
kubectl rollout restart deployment/app-receiver -n production

# - If using Docker Swarm, update service
docker service update --force-update app-receiver

# Option 3: Wait for automatic cache invalidation
# Receivers will re-fetch JWKS when cache expires (default 60s)
# Max propagation time: 60 seconds
```

**Step 7: Monitor JWKS endpoint**
```bash
# Verify new public key is in JWKS
curl https://issuer.example.com/.well-known/jwks.json | jq '.'

# Should show:
# {
#   "keys": [
#     {
#       "kty": "OKP",
#       "crv": "Ed25519",
#       "x": "<NEW_KEY_MATERIAL>",
#       "kid": "<NEW_KEY_ID>",
#       "use": "sig"
#     }
#   ]
# }

# Verify old key is NOT present
curl https://issuer.example.com/.well-known/jwks.json | jq '.keys[] | .kid'
```

### Verification (10+ minutes)

**Step 8: Verify tokens work with new key**
```bash
# Create test token with new key
curl -X POST http://localhost:4000/token \
  -H "Content-Type: application/json" \
  -d '{"subject":"svc://test","claims":{"test":"true"}}'

# Verify receiver accepts it
curl http://localhost:4001/protected \
  -H "Authorization: Bearer <NEW_TOKEN>"

# Should return 200 (authorized)

# Verify old tokens are rejected (or accepted with short TTL remaining)
# Depends on if old key is still in JWKS or removed
```

**Step 9: Check all receivers are updated**
```bash
# For each receiver endpoint, verify it accepts new tokens:
for receiver in receiver1.example.com receiver2.example.com receiver3.example.com; do
  curl https://$receiver/healthz
  # Verify 200 response
done
```

### Post-Incident (After recovery)

**Step 10: Incident forensics**
```bash
# 1. Archive audit logs for investigation
tar -czf /backups/oathmesh-audit-2024-XX-XX.tar.gz /var/log/oathmesh/

# 2. Check for suspicious token usage
grep -i "verify.*success" /var/log/oathmesh/audit.log | \
  jq '.subject' | sort | uniq -c | sort -rn

# 3. Identify affected services
# (services that have been authenticated to during compromise window)
```

**Step 11: Root cause analysis**
- How was the key leaked?
- What was the exposure window?
- Were any malicious requests made?
- Which services/subjects were affected?

**Step 12: Remediation**
- [ ] Rotate secrets on all affected services
- [ ] Review IAM permissions for key access
- [ ] Update secrets management procedures
- [ ] Schedule post-mortem meeting within 24 hours

### Communication Template

```
[INCIDENT] Private Key Compromise

Status: RESOLVED
Detection: 2024-XX-XX 14:23:00 UTC
Resolution: 2024-XX-XX 14:31:00 UTC

Summary:
- Private key was compromised via [CAUSE]
- Compromised from 2024-XX-XX HH:MM to 2024-XX-XX HH:MM UTC
- New key deployed and propagated to all receivers
- No legitimate services were impacted

Impact:
- [NUMBER] suspicious token validation attempts detected
- [NUMBER] services affected (none compromised)
- ~10 minutes downtime during key rotation

Next Steps:
- Post-mortem scheduled for [DATE/TIME]
- Key rotation procedures to be updated
- Secrets management audit planned

Contact: [SECURITY_LEAD] for details
```

---

## Incident: Issuer Service Down

**Severity:** 🟠 **HIGH** (if brief) or 🔴 **CRITICAL** (if prolonged)  
**Detection Time:** < 1 minute  
**Recovery Time:** < 30 minutes  
**Owner:** Backend Lead + SRE

### Detection Signals

1. **Issuer health check failing**
   ```bash
   # If monitored:
   Alert: "oathmesh_issuer_health_check failed"
   ```

2. **Manual health check**
   ```bash
   curl http://localhost:4000/healthz
   # Returns: Connection refused or 5xx error
   ```

3. **Token issuance failing**
   ```bash
   curl -X POST http://localhost:4000/token \
     -H "Content-Type: application/json" \
     -d '{"subject":"svc://test"}'
   # Returns: 500, 502, or connection refused
   ```

### Immediate Response (0-5 minutes)

**Step 1: Check service status**
```bash
# Docker
docker ps | grep oathmesh-issuer
# Should show container running with recent start time

# Or systemd
systemctl status oathmesh-issuer

# Or Kubernetes
kubectl get pods -l app=oathmesh-issuer -n production
# Should show Running status
```

**Step 2: Check logs for errors**
```bash
# Docker
docker logs oathmesh-issuer -n 100
# Look for: panic, error, fatal

# Or
journalctl -u oathmesh-issuer -n 100
```

**Step 3: Check dependencies**
```bash
# Redis connectivity
redis-cli ping
# Returns: PONG

# Database (if applicable)
psql -h localhost -U oathmesh -c "SELECT 1"
# Returns: 1

# Network (if issuer uses external API)
curl http://external-service/healthz
```

**Step 4: Restart the service (if no critical issue found)**
```bash
# Docker
docker restart oathmesh-issuer

# Systemd
systemctl restart oathmesh-issuer

# Kubernetes
kubectl rollout restart deployment/oathmesh-issuer -n production
```

**Step 5: Verify recovery**
```bash
# Wait 10 seconds for startup
sleep 10

# Check health
curl http://localhost:4000/healthz
# Returns: 200 OK

# Test token creation
curl -X POST http://localhost:4000/token \
  -H "Content-Type: application/json" \
  -d '{"subject":"svc://test","claims":{"env":"prod"}}'
# Returns: 200 with token
```

### Common Failure Modes

| Symptom | Cause | Fix |
|---------|-------|-----|
| `ERROR: connection refused` | Service not running | `docker start` or `systemctl start` |
| `ERROR: JWKS verification failed` | Redis down or key corrupted | Check Redis, restart both issuer+Redis |
| `ERROR: open config: no such file` | Config file deleted/moved | Restore config, verify permissions |
| `ERROR: OOM` | Memory leak or high load | Restart service, monitor memory usage |
| `ERROR: panic in policy engine` | Malformed policy | Rollback to previous policy, check syntax |

### Fallback Behavior (During Outage)

**While issuer is down:**
- ✅ Existing tokens still work (receivers have cached JWKS)
- ✅ Receivers cache JWKS for up to 60 seconds
- ❌ New tokens cannot be issued
- ❌ Revocation updates may be stale

**Maximum safe outage:** 5 minutes
- After 5 min, some tokens will start expiring without issuance
- After 60 min, all cached JWKS will expire

### Recovery Confirmation

```bash
# 1. Issuer health
curl http://localhost:4000/healthz
# Returns: 200

# 2. Can issue tokens
TOKEN=$(curl -s -X POST http://localhost:4000/token \
  -H "Content-Type: application/json" \
  -d '{"subject":"svc://test"}' | jq -r '.token')
echo $TOKEN

# 3. Receivers can verify tokens
curl http://localhost:4001/protected \
  -H "Authorization: Bearer $TOKEN"
# Returns: 200 (authorized)

# 4. Revocation list is syncing
curl http://localhost:4000/metrics | grep revocation_sync
# Look for: oathmesh_revocation_sync_errors 0 (no recent errors)

# 5. No recent errors in logs
docker logs oathmesh-issuer | tail -20
# Should be clean (no errors in last few lines)
```

---

## Incident: Redis Revocation List Down

**Severity:** 🟠 **HIGH**  
**Detection Time:** < 1 minute  
**Recovery Time:** < 5 minutes  
**Owner:** SRE + Backend Lead

### Detection Signals

1. **Alert: RevocationSyncFailures**
   ```
   Alert firing: "oathmesh_revocation_sync_errors rate > 0"
   ```

2. **Manual check**
   ```bash
   redis-cli ping
   # Returns: PONG (if Redis healthy)
   # Returns: Connection refused (if Redis down)
   ```

3. **Issuer logs show Redis errors**
   ```bash
   docker logs oathmesh-issuer | grep -i redis
   # ERROR: redis: connection refused
   # ERROR: revocation sync failed: redis timeout
   ```

### Immediate Response (0-5 minutes)

**Step 1: Verify Redis is actually down**
```bash
# Try to connect
redis-cli ping

# If no response after 5 seconds, Redis is down
```

**Step 2: Check Redis status**
```bash
# Docker
docker ps | grep redis
# If not in list, container is stopped

docker logs oathmesh-redis -n 50
# Look for: panic, fatal error, OOM killer

# Or systemd
systemctl status redis-server
```

**Step 3: Check common failure causes**
```bash
# Disk full?
df -h | grep var

# Memory exhausted?
free -h
# or
docker stats oathmesh-redis

# Port already in use?
netstat -tuln | grep 6379

# Permissions issue?
ls -la /var/lib/redis
# Should show redis:redis ownership
```

**Step 4: Restart Redis**
```bash
# Docker
docker restart oathmesh-redis

# Systemd
systemctl restart redis-server

# Wait for startup
sleep 5

# Verify it started
redis-cli ping
# Returns: PONG
```

**Step 5: Verify issuer is re-syncing**
```bash
# Check issuer logs
docker logs oathmesh-issuer -n 50 | grep -i revocation

# Should see: "revocation list synced" or similar
# Should NOT see: "revocation sync failed"

# Check metrics
docker exec oathmesh-issuer curl localhost:9090/metrics | grep revocation_sync
# Look for: oathmesh_revocation_sync_errors 0 (no recent errors after restart)
```

### Impact During Outage

**While Redis is down:**
- ✅ Token verification still works (revocation checks skipped or use in-memory list)
- ✅ Issuer continues operating
- ⚠️ **Revocation list is stale** (no updates possible)
- ⚠️ Recently revoked subjects might still authenticate

**Maximum safe outage:** 15 minutes
- After 15 min, unrevoked subjects may authenticate despite revocation request
- After 1 hour, in-memory cache may have stale data

### Recovery Verification

```bash
# 1. Redis is responsive
redis-cli ping
# Returns: PONG

# 2. Can read revocation data
redis-cli HGETALL oathmesh:revocations
# Returns hash of revoked subjects (or empty if none)

# 3. Issuer is syncing successfully
docker logs oathmesh-issuer | tail -5
# No revocation sync errors

# 4. Metrics show recovery
docker exec oathmesh-issuer curl -s localhost:9090/metrics | \
  grep oathmesh_revocation_sync_errors
# Should show: oathmesh_revocation_sync_errors 0

# 5. Revocation still enforced
# Try to verify a revoked token - should be rejected
```

### Post-Incident (After recovery)

**Step 6: Investigate root cause**
```bash
# Why did Redis crash?
docker logs oathmesh-redis | grep -E "(panic|fatal|oom-killer|segmentation)"

# Was it a crash or intentional stop?
docker inspect oathmesh-redis | grep -A5 '"State"'
# ExitCode -1 = crashed
# ExitCode 0 = stopped gracefully

# Memory usage patterns?
# (check if trend shows growth toward limit)
```

**Step 7: Prevent recurrence**
- [ ] Set Redis memory limit and eviction policy
- [ ] Add Redis memory usage monitoring alert
- [ ] Increase Redis persistent storage monitoring
- [ ] Review if Redis configuration is appropriate for workload

---

## Reference Materials

- [OathMesh Architecture](../ARCHITECTURE.md)
- [Monitoring & Observability](./monitoring.md)
- [On-Call Runbook](./on-call-runbook.md)
- [Troubleshooting Guide](./troubleshooting.md)
