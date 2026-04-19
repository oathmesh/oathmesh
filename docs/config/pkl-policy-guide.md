# Pkl Policy Guide

## Overview

OathMesh uses Apple Pkl for policy configuration. Pkl provides compile-time schema validation, type safety, and clear error messages for policy authors.

## The Policy Lifecycle

```
Author → Validate → Deploy → Hot-Reload
```

### 1. Author

Create a `.pkl` file that amends `schema.pkl`:

```pkl
amends "schema.pkl"

version = 1

issuers {
  "https://issuer.oathmesh.tech"
}

audiences {
  "https://inventory.internal"
}

rules {
  new {
    name = "storefront-read"
    match {
      sub = "agent://repo/acme/*"
      act = "inventory.read"
    }
    allow = true
  }
  new {
    name = "deploy-write"
    match {
      sub = "agent://repo/acme/deploy-bot"
      act = "inventory.write"
      src {
        type = "github_actions"
        repo = "acme/storefront"
        workflow = "deploy.yml"
      }
    }
    allow = true
  }
  // Default deny — REQUIRED. Never remove. Never set allow = true.
  new {
    name = "default"
    allow = false
  }
}
```

### 2. Validate

```bash
oathmesh policy validate policy/production.pkl
```

Or using Pkl directly:

```bash
pkl eval policy/production.pkl
```

Schema errors surface at validation time with line numbers and clear messages — not at request time.

### 3. Deploy

Place the policy file where the issuer or gateway can read it. Set the path via environment variable or CLI flag:

```bash
OATHMESH_GATEWAY_POLICY=policy/production.pkl
```

### 4. Hot-Reload

OathMesh watches the policy file via `fsnotify`. When the file changes on disk:

1. The new policy is loaded and validated against the schema
2. If valid: atomic swap — zero downtime
3. If invalid: the old policy remains active, an error is logged

No server restart required.

## Rule Evaluation

- Rules are evaluated **in order** — first match wins
- An empty match (`match {}`) matches all requests
- The last rule **must** be `{ name = "default", allow = false }`
- This default-deny rule cannot be removed or set to `allow = true`

## Match Criteria

| Field | Type | Matching |
|---|---|---|
| `sub` | string | Glob pattern (e.g., `agent://repo/acme/*`). `*` matches any characters including `/`. |
| `act` | string | Exact match |
| `scope` | string[] | All listed scope values must be present in the token |
| `env` | string | Exact match |
| `src.type` | string | Exact match |
| `src.repo` | string | Glob pattern |
| `src.workflow` | string | Glob pattern |

All specified fields must match (AND logic). Unspecified fields are wildcards.

## Scope Matching

Scope matching requires the token's scope to be a **superset** of the rule's scope. If a rule specifies `scope: ["inventory.read", "inventory.write"]`, the token must contain **both** values.

## See Also

- [Policy Overview](../policies/overview.md)
- [Policy Examples](../policies/examples.md)
- [Policy Migration Guide](../policies/migration.md)
