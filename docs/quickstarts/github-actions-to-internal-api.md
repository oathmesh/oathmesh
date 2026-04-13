# Quickstart: GitHub Actions to Internal API

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

<p align="center">
  <b>Authenticate GitHub Actions workflows against internal APIs — zero long-lived secrets.</b>
</p>

<p align="center">
  <b>⏱️ Time:</b> ~15 minutes
</p>

---

> 🆕 **New here?** Start with the [Quick Start](../README.md#-quick-start) in the main README.

This guide shows how a GitHub Actions workflow can authenticate against an internal API protected by OathMesh — zero long-lived secrets.

## How It Works

1. GitHub Actions issues an OIDC token for your workflow run
2. Your workflow exchanges that OIDC token with the OathMesh issuer for an Oath Token
3. Your workflow calls the internal API with the Oath Token
4. The API verifies the token and processes the request

## Prerequisites

- A running OathMesh issuer accessible from GitHub Actions runners (public URL or self-hosted runner)
- Repository configured with `id-token: write` permission

## Step 1: Configure Your Workflow

```yaml
name: Deploy Sync
on: push

jobs:
  sync:
    runs-on: ubuntu-latest
    permissions:
      id-token: write    # Required for OIDC token
      contents: read

    steps:
      - name: Get GitHub OIDC Token
        id: oidc
        run: |
          TOKEN=$(curl -s -H "Authorization: bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
            "$ACTIONS_ID_TOKEN_REQUEST_URL&audience=https://issuer.oathmesh.dev" \
            | jq -r '.value')
          echo "GH_TOKEN=$TOKEN" >> $GITHUB_ENV

      - name: Exchange for Oath Token
        run: |
          RESPONSE=$(curl -s -X POST "https://issuer.oathmesh.dev/v1/exchange/github" \
            -H "Content-Type: application/json" \
            -d '{"github_token": "'"$GH_TOKEN"'"}')
          
          OATH_TOKEN=$(echo $RESPONSE | jq -r '.token')
          echo "OATH_TOKEN=$OATH_TOKEN" >> $GITHUB_ENV

      - name: Call Internal API
        run: |
          curl -H "Authorization: OathMesh $OATH_TOKEN" \
            https://inventory.internal/sync
```

## Step 2: Issuer Configuration

The issuer automatically maps GitHub OIDC claims to OathMesh source provenance:

| GitHub Claim | OathMesh Mapping |
|---|---|
| `repository` | `src.repo` |
| `workflow` | `src.workflow` |
| `run_id` | `src.run_id` |
| `sha` | `src.sha` |

The subject is auto-derived as: `job://github/{repo}/{workflow}`

## Step 3: Policy Configuration

Create a policy rule that allows your workflow:

```pkl
new {
  name = "deploy-sync"
  match {
    sub = "job://github/acme/storefront/*"
    act = "inventory.write"
    src {
      type = "github_actions"
      repo = "acme/storefront"
    }
  }
  allow = true
}
```

## Security Notes

- The GitHub OIDC token is verified against GitHub's published JWKS **before** any exchange processing begins
- The resulting Oath Token is short-lived (≤300s) — it cannot be stored and reused across workflow runs
- Each workflow run gets a unique token with a unique `jti`

## Next Steps

- [Run the full demo locally](local-demo-docker-compose.md)
- [Policy configuration guide](../config/pkl-policy-guide.md)
- [Protect a Go chi API](protect-chi-api.md)
- [Protect an Express API](protect-express-api.md)
- [Protect a Next.js API](protect-nextjs-api.md)
- [Protect a FastAPI service](protect-fastapi.md)

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [Protocol: Token Exchange](link-to-be-added) | OIDC exchange flow details |
| [Verification Rules](../docs/protocol/verification-rules.md) | 14-step pipeline details |
| [Error Taxonomy](../docs/protocol/error-taxonomy.md) | All error codes and meanings |
| [Security: Threat Model](../docs/security/threat-model.md) | Security model |
| [Config: Pkl Policy Guide](../config/pkl-policy-guide.md) | Policy DSL reference |
| [Protocol: Source Provenance](../docs/protocol/claim-reference.md) | `src` claim details |
