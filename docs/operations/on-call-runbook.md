# On-Call Runbook

**Last Updated:** 2024 | **Status:** Active  
**Owner:** SRE Team | **Response Time Targets:** See below

---

## Getting Started (First 5 Minutes)

### Step 1: Acknowledge the Alert

```bash
# In your alert system (PagerDuty, Opsgenie, etc.):
1. Click "Acknowledge" to let team know you're handling it
2. Read the alert title and description
3. Note the severity (🔴 Critical / 🟠 High / 🟡 Medium)
```

### Step 2: Gather Status Information

Run this health check script:

```bash
#!/bin/bash
echo "=== OathMesh Health Check ==="

# 1. Issuer status
echo -e "\n[1] ISSUER STATUS"
docker ps | grep oathmesh-issuer || echo "❌ Issuer not running"
curl -s http://localhost:4000/healthz && echo "✅ Issuer healthy" || echo "❌ Issuer unreachable"

# 2. Redis status
echo -e "\n[2] REDIS STATUS"
docker ps | grep redis || echo "❌ Redis not running"
redis-cli ping && echo "✅ Redis healthy" || echo "❌ Redis unreachable"

# 3. Recent errors
echo -e "\n[3] RECENT ERRORS (Last 50 lines)"
docker logs oathmesh-issuer -n 50 | grep -i error || echo "✅ No recent errors"

# 4. Metrics check
echo -e "\n[4] KEY METRICS"
curl -s http://localhost:9090/metrics | grep -E "(revocation_sync_errors|verification_errors|http_requests)" | head -5

# 5. Disk space
echo -e "\n[5] DISK SPACE"
df -h / | tail -1

echo -e "\n=== End Health Check ==="
```

### Step 3: Determine Severity

| Signal | Severity | Next Step |
|--------|----------|-----------|
| Issuer not running + Redis OK | 🟠 High | Restart issuer |
| Redis not running + Issuer OK | 🟠 High | Restart Redis |
| Both down | 🔴 Critical | Restart both, then escalate |
| Errors in logs but services running | 🟡 Medium | Investigate error type |
| No errors, all services running | 🟢 Low | Check if alert is false positive |

---

## Common Issues & Quick Fixes

### Issue 1: "Connection Refused" Error

**Symptom:** `curl http://localhost:4000/healthz` returns connection refused

**Fix (1 minute):**
```bash
# Check if issuer is running
docker ps | grep oathmesh-issuer
# If empty → not running

# Restart it
docker restart oathmesh-issuer
docker logs oathmesh-issuer -n 20

# Wait 5 seconds, test again
sleep 5
curl http://localhost:4000/healthz
```

### Issue 2: "Redis Connection Failed"

**Symptom:** Logs show `ERROR: redis: connection refused`

**Fix (1 minute):**
```bash
# Test Redis connection
redis-cli ping
# If "PONG" → Redis is fine, might be network issue

# If no response → Redis is down
docker ps | grep redis
# If empty → not running

# Restart it
docker restart oathmesh-redis
redis-cli ping  # Verify it came back up
```

### Issue 3: "Revocation Sync Failed"

**Symptom:** Alert `oathmesh_revocation_sync_errors` firing

**Fix (2 minutes):**
```bash
# Check if Redis is accessible
redis-cli ping

# Check issuer Redis connection URL
docker exec oathmesh-issuer env | grep REDIS
# Should show REDIS_URL set correctly

# Check recent errors
docker logs oathmesh-issuer -n 100 | grep -i "revocation"

# If Redis is fine, try restarting issuer
docker restart oathmesh-issuer
sleep 5
docker logs oathmesh-issuer | tail -10
# Should show revocation list synced successfully
```

### Issue 4: "High Verification Error Rate"

**Symptom:** Alert `oathmesh_verification_errors` increased significantly

**Possible Causes & Fixes:**

```bash
# A) JWKS cache expired (receivers couldn't refresh)
# Fix: Check issuer health and restart receivers
docker ps | grep oathmesh-issuer
curl http://localhost:4000/.well-known/jwks.json

# B) Recent key rotation (old key no longer valid)
# Fix: Clear receiver caches
curl -X POST http://localhost:4001/internal/cache/clear

# C) Receiver misconfiguration
# Fix: Check receiver logs
docker logs app-receiver -n 50 | grep -i error

# D) Network issue between receiver and issuer
# Fix: Check connectivity
docker exec app-receiver ping oathmesh-issuer
```

### Issue 5: "Disk Space Critically Low"

**Symptom:** `df -h /` shows usage > 90%

**Fix (5 minutes):**
```bash
# 1. Find what's using space
du -sh /* | sort -rh | head -10

# 2. Check container logs (often the culprit)
du -sh /var/lib/docker/containers/* | sort -rh | head -5

# 3. Clean up old logs (safe to do)
docker system prune -a  # WARNING: removes unused images/containers
# or more conservative:
docker container prune -f
docker image prune -f

# 4. If still full, check database/Redis data
redis-cli INFO memory  # Check Redis memory usage
du -sh /var/lib/redis/
```

### Issue 6: "Memory Usage Too High"

**Symptom:** Docker shows memory usage > 90% of limit

**Fix (3 minutes):**
```bash
# Check what's using memory
docker stats --no-stream | sort -k4 -rh

# If issuer is high:
docker restart oathmesh-issuer

# If Redis is high:
# This is normal if revocation list is large
redis-cli INFO memory  # Check stats

# If Redis memory keeps growing, possible memory leak:
redis-cli MEMORY DOCTOR  # Get Redis memory analysis
docker restart oathmesh-redis
```

---

## Decision Tree: What To Do Next?

```
START
├─ Is issuer health check passing?
│  ├─ NO → [Issue 1] Restart issuer
│  └─ YES → Continue
│
├─ Is Redis health check passing?
│  ├─ NO → [Issue 2] Restart Redis
│  └─ YES → Continue
│
├─ Is revocation sync working?
│  ├─ NO → [Issue 3] Restart issuer
│  └─ YES → Continue
│
├─ Are verification errors high?
│  ├─ YES → [Issue 4] Check JWKS/cache
│  └─ NO → Continue
│
├─ Is disk space low?
│  ├─ YES → [Issue 5] Clean up
│  └─ NO → Continue
│
├─ Is memory usage high?
│  ├─ YES → [Issue 6] Investigate service
│  └─ NO → Continue
│
└─ Alert is likely false positive or intermittent issue
   └─ Monitor for 5 min, clear if no recurrence
```

---

## Escalation Guide

### When to Escalate

**Escalate to Backend Lead if:**
- Restart didn't fix the issue
- Multiple services are down simultaneously
- Logs show application logic errors (not just connection errors)
- You're unsure what to do after 10 minutes

**Escalate to Security Lead if:**
- Alert involves authentication failure
- Suspicious activity in audit logs
- Potential key compromise

**Escalate to CTO if:**
- Multiple teams affected
- Public status update needed
- Major incident (> 30 min downtime)

### Escalation Template

```
[ESCALATION] Issue Name

Severity: 🔴 Critical / 🟠 High / 🟡 Medium

What I've tried:
1. [Action 1] - Result: [Result]
2. [Action 2] - Result: [Result]

Current status:
- [Service] - [Status]
- [Service] - [Status]

Why escalating:
- [Reason]

What I need:
- [Specific ask]
```

---

## Monitoring Dashboard Quick Reference

### Key Metrics to Watch

**Prometheus metrics URL:** `http://localhost:9090/graph`

**Top 5 metrics to check:**

1. **`oathmesh_verification_errors`**
   - What: Number of token verification failures
   - Normal: 0-10/min
   - Alert: > 100/min for 5 min

2. **`oathmesh_revocation_sync_errors`**
   - What: Revocation list sync failures
   - Normal: 0
   - Alert: > 0 (any failure)

3. **`oathmesh_http_requests_duration_seconds`**
   - What: Request latency
   - Normal: < 100ms p99
   - Alert: > 500ms p99

4. **`oathmesh_issuer_uptime_seconds`**
   - What: How long issuer has been running
   - Normal: > 86400 (1 day)
   - Alert: < 300 (recently restarted)

5. **`redis_connected_clients`**
   - What: Number of Redis connections
   - Normal: 1-5
   - Alert: > 100 (connection leak?)

---

## Log File Locations & Parsing

### Finding Logs

```bash
# Issuer application logs
docker logs oathmesh-issuer -n 100
docker logs oathmesh-issuer -f  # Follow in real-time

# Redis logs
docker logs oathmesh-redis -n 50

# System logs (if running on host)
journalctl -u oathmesh-issuer -n 100
journalctl -u redis-server -n 50

# Audit logs (if configured)
tail -f /var/log/oathmesh/audit.log
```

### Parsing Key Information

```bash
# Find all errors
docker logs oathmesh-issuer | grep -i error

# Find panic/crash
docker logs oathmesh-issuer | grep -i panic

# Find Redis connection issues
docker logs oathmesh-issuer | grep -i redis

# Count errors by type
docker logs oathmesh-issuer -n 1000 | grep -i error | sort | uniq -c | sort -rn

# Get summary of last 10 lines
docker logs oathmesh-issuer -n 10 | tail -10
```

---

## Testing & Verification

### Health Check Commands

```bash
# Comprehensive health check
curl -s http://localhost:4000/healthz | jq '.'
# Should return: {"status":"ok"}

# Can issue tokens?
curl -s -X POST http://localhost:4000/token \
  -H "Content-Type: application/json" \
  -d '{"subject":"svc://test"}' | jq '.token' | head -c 50

# Can verify tokens?
TOKEN=$(curl -s -X POST http://localhost:4000/token \
  -H "Content-Type: application/json" \
  -d '{"subject":"svc://test"}' | jq -r '.token')
curl -s http://localhost:4001/protected -H "Authorization: Bearer $TOKEN"

# JWKS endpoint working?
curl -s http://localhost:4000/.well-known/jwks.json | jq '.keys | length'
# Should return: 1 (or number of keys in rotation)
```

### Metrics Export Check

```bash
# Are metrics being exported?
curl -s http://localhost:9090/metrics | head -20

# Look for oathmesh metrics
curl -s http://localhost:9090/metrics | grep oathmesh | head -10
```

---

## Reference Materials

### Quick Links

- **Incident Response Playbook:** `docs/operations/incident-response.md`
- **Architecture Docs:** `docs/ARCHITECTURE.md`
- **Configuration Guide:** `docs/operations/configuration.md`
- **Troubleshooting Guide:** `docs/operations/troubleshooting.md`

### Emergency Contacts

| Role | Contact | Availability |
|------|---------|---|
| Backend Lead | [Name] | 24/7 on-call |
| Security Lead | [Name] | 24/7 on-call |
| DevOps Lead | [Name] | Business hours |
| CTO | [Name] | Escalations only |

---

## After-Hours Support

### SLA Response Times

- 🔴 **Critical:** 5 minutes
- 🟠 **High:** 15 minutes
- 🟡 **Medium:** 1 hour

### During Incident

1. **Acknowledge** alert immediately (shows you're responding)
2. **Update status** every 5 minutes if unresolved
3. **Escalate** if unsure after 10 minutes
4. **Communicate** with team once resolved

### Post-Incident

- Document what happened in incident report
- Schedule post-mortem if severity ≥ High
- Update runbook with new findings
- Share learnings with team

---

## Getting Help

### Questions During On-Call?

1. Check this runbook
2. Search recent incident reports
3. Ask Backend Lead (escalation)
4. Check #oathmesh-incidents Slack channel

### Runbook Improvements?

Found something that needs updating?
1. Note the issue
2. Update this file after incident is resolved
3. Notify team in post-mortem
4. Test new procedure next drill

---

**Remember: You're not alone. This is what the on-call rotation is for.**
**When in doubt, escalate. It's better to ask for help than to make it worse.**
