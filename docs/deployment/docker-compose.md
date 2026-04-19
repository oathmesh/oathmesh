← [Back to Index](../INDEX.md)

# Docker Compose Deployment (Production-Style Local/Self-Hosted)

This guide uses `docker-compose.prod.yml` to run:

- OathMesh issuer
- Redis (persistent + password protected)
- Optional sample receiver (`sample-receiver` profile)

## 1) Prerequisites

- Docker Engine + Docker Compose v2
- Ed25519 private key file (example: `private.pem`)

Generate a key if needed:

```bash
openssl genpkey -algorithm Ed25519 -out private.pem
```

## 2) Configure secrets and environment

Create `.env` at repo root (do **not** commit it):

```dotenv
# Required
OATHMESH_ISSUER=https://issuer.local.example
OATHMESH_MINT_SECRET=replace-with-long-random-secret
REDIS_PASSWORD=replace-with-long-random-secret
OATHMESH_PRIVATE_KEY_PATH=./private.pem

# Optional overrides
ISSUER_PORT=4000
SAMPLE_RECEIVER_PORT=8081
SAMPLE_RECEIVER_AUDIENCE=https://inventory.internal
SAMPLE_RECEIVER_TRUSTED_ISSUERS=http://issuer:4000
OATHMESH_AUDIT_SINK=stdout
```

> Critical: keep real secrets in `.env`, your shell environment, or a secret manager; never hardcode them in YAML.

## 3) Start services

Issuer + Redis:

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

Include sample receiver:

```bash
docker compose -f docker-compose.prod.yml --profile sample-receiver up -d --build
```

## 4) Verify deployment

```bash
# Issuer liveness
curl -f http://localhost:4000/healthz

# Service health and status
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs --tail=100 issuer redis

# Validate config rendering
docker compose -f docker-compose.prod.yml config
```

If sample receiver is enabled:

```bash
curl -f http://localhost:8081/healthz
```

## 5) Operational commands

```bash
# Restart a single service
docker compose -f docker-compose.prod.yml restart issuer

# Pull/rebuild and roll forward
docker compose -f docker-compose.prod.yml up -d --build

# Stop stack (keep data)
docker compose -f docker-compose.prod.yml down

# Stop stack and remove Redis persisted data
docker compose -f docker-compose.prod.yml down -v
```

## 6) Troubleshooting

### Issuer fails with key/secret errors
- Confirm `.env` exists and contains `OATHMESH_MINT_SECRET` and `OATHMESH_PRIVATE_KEY_PATH`.
- Verify the key file exists and is readable at the configured path.

### Issuer cannot use Redis
- Check `REDIS_PASSWORD` matches for both `redis` and `REDIS_URL` (auto-wired in compose).
- Inspect logs:
  ```bash
  docker compose -f docker-compose.prod.yml logs --tail=200 redis issuer
  ```

### Healthcheck stays `unhealthy`
- Confirm ports are free (`4000`, optional `8081`).
- Re-run with fresh build:
  ```bash
  docker compose -f docker-compose.prod.yml down
  docker compose -f docker-compose.prod.yml up -d --build
  ```

### Config interpolation errors (`set ... in .env`)
- A required environment variable is missing. Add it to `.env` and rerun.

## Related

- [TLS Deployment Guide](./tls.md)

