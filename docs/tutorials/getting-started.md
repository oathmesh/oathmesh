---
title: Getting Started Tutorial
description: Run an end-to-end local issuer and receiver flow, mint a token, verify it, and call a protected API.
tags: [tutorial, getting-started, issuer, receiver, verification]
---

← [Back to Index](../INDEX.md)

# Tutorial: Getting Started (Issuer + Receiver + Verification)

Run a real local flow: start issuer + receiver, mint a token, verify it, and call a protected API.

## Prerequisites

- Docker + Docker Compose
- Go 1.26.2+
- `curl` and `jq`

## 1) Start local services

From repo root:

```bash
docker compose up -d issuer chi-api redis
```

Health checks:

```bash
curl -s http://localhost:4000/healthz
curl -s http://localhost:8081/healthz
```

Expected output for both: `OK`

## 2) Mint a token from the issuer

`/v1/token` is protected by `OATHMESH_MINT_SECRET` (in `docker-compose.yml` this is `development_secret_do_not_use_in_prod`).

```bash
TOKEN=$(curl -s http://localhost:4000/v1/token \
  -X POST \
  -H "Authorization: Bearer development_secret_do_not_use_in_prod" \
  -H "Content-Type: application/json" \
  -d '{
    "sub": "svc://local/tutorial-client",
    "aud": "https://inventory.internal",
    "act": "inventory.read",
    "ttl_hint": 120
  }' | jq -r '.token')
```

Quick check:

```bash
test -n "$TOKEN" && echo "token minted"
```

Expected output: `token minted`

## 3) Verify the token with CLI

Build CLI once:

```bash
go build -o bin/oathmesh ./cmd/oathmesh
echo "$TOKEN" | ./bin/oathmesh verify \
  --audience "https://inventory.internal" \
  --issuer "http://localhost:4000"
```

Expected output includes:

- `✅ Token verified successfully`
- `Subject: svc://local/tutorial-client`

## 4) Call the protected receiver

```bash
curl -s -H "Authorization: OathMesh $TOKEN" http://localhost:8081/inventory | jq .
```

Expected output includes:

- `"status": "success"`
- `"subject": "svc://local/tutorial-client"` (inside `caller.principal`)

## 5) Verify replay protection (real deny path)

Reuse the same token:

```bash
curl -s -i -H "Authorization: OathMesh $TOKEN" http://localhost:8081/inventory
```

Expected result: `HTTP/1.1 401` with `replay_detected`.

## Exchange flow

For GitHub/GitLab OIDC exchange (`POST /v1/exchange/github`, `POST /v1/exchange/gitlab`) see:

- [CI/CD machine identity tutorial](ci-cd-machine-identity.md)

## Troubleshooting

- `mint_auth_required` / `mint_auth_denied`: wrong or missing `Authorization: Bearer <OATHMESH_MINT_SECRET>` on `/v1/token`.
- `issuer_untrusted`: verify command or receiver trusted issuers do not include `http://localhost:4000`.
- `audience_mismatch`: minted `aud` must be `https://inventory.internal` for `examples/chi-api`.
- `connection refused`: `docker compose ps` and confirm `issuer` (4000) and `chi-api` (8081) are up.

## Cleanup

```bash
docker compose down
```

## Related docs

- [Docker quickstart](../quickstarts/local-demo-docker-compose.md)
- [CI/CD machine identity tutorial](ci-cd-machine-identity.md)
- [Token format reference](../protocol/token-format.md)
