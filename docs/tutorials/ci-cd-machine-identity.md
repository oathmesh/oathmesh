← [Back to Index](../INDEX.md)

# Tutorial: CI/CD Machine Identity (GitHub Actions Pattern)

Use GitHub OIDC + OathMesh exchange/mint endpoints without storing long-lived API credentials in workflow code.

## 1) Workflow skeleton

```yaml
name: OathMesh CI Call
on: [push]

jobs:
  call-internal-api:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    env:
      OATHMESH_ISSUER: https://issuer.example.com
      OATHMESH_MINT_SECRET: ${{ secrets.OATHMESH_MINT_SECRET }}
    steps:
      - name: Request GitHub OIDC token
        run: |
          GH_TOKEN=$(curl -s -H "Authorization: bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
            "${ACTIONS_ID_TOKEN_REQUEST_URL}&audience=${OATHMESH_ISSUER}" | jq -r '.value')
          echo "GH_TOKEN=$GH_TOKEN" >> $GITHUB_ENV

      - name: Exchange GitHub token (issuer verifies GitHub JWT)
        run: |
          EXCHANGE=$(curl -s -X POST "${OATHMESH_ISSUER}/v1/exchange/github" \
            -H "Authorization: Bearer ${OATHMESH_MINT_SECRET}" \
            -H "Content-Type: application/json" \
            -d "{\"github_token\":\"${GH_TOKEN}\"}")
          echo "$EXCHANGE" | jq .

      - name: Mint API call token for receiver audience
        run: |
          API_TOKEN=$(curl -s -X POST "${OATHMESH_ISSUER}/v1/token" \
            -H "Authorization: Bearer ${OATHMESH_MINT_SECRET}" \
            -H "Content-Type: application/json" \
            -d "{\"sub\":\"job://github/${{ github.repository }}/${{ github.workflow }}\",\"aud\":\"https://inventory.internal\",\"act\":\"inventory.sync\",\"ttl_hint\":120}" \
            | jq -r '.token')
          echo "API_TOKEN=$API_TOKEN" >> $GITHUB_ENV

      - name: Call protected API
        run: |
          curl -sS -H "Authorization: OathMesh ${API_TOKEN}" https://inventory.internal/sync
```

## Why this matches current repo behavior

- Exchange endpoint exists: `POST /v1/exchange/github` (`internal/issuer/server.go`)
- Exchange request body key is `github_token` (`internal/issuer/exchange.go`)
- Mint endpoint exists: `POST /v1/token` and both are protected by `MintAuth` (`Authorization: Bearer <OATHMESH_MINT_SECRET>`)

## Security notes

- Keep `OATHMESH_MINT_SECRET` only in runner secret storage.
- Never print OIDC or Oath tokens in logs.
- Keep `ttl_hint` short (<=300s is enforced by signer).
- Pin receiver-side `aud` and `trusted_issuers`.
- Treat exchange output as issuer-attested job identity; mint receiver-specific token for API calls.
