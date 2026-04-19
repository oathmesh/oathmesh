# Policy Overview

This guide documents the current Go policy engine behavior (`internal/policy`) used during verification Step 14.

## Policy model

- Rules are evaluated **top to bottom**.
- **First match wins** (allow or deny).
- If no rule matches, result is **deny**.
- A valid policy must end with:

```pkl
new {
  name = "default"
  allow = false
}
```

- All fields inside `match` are combined with **AND**.
- Omitted fields are wildcards.

## Matching semantics

Supported `match` fields:

- `sub`
- `act`
- `scope`
- `env`
- `tenant`
- `src.type`, `src.repo`, `src.workflow`

Field behavior:

- `scope` is superset matching: token scope must include **all** values listed in rule scope.
- `env` and `tenant` are exact match.
- `sub`, `act`, and `src.*` use simple glob matching with `*`:
  - `*` (match all)
  - `prefix*`
  - `*suffix`
  - `pre*suf`
  - exact string (no wildcard)

> Only one wildcard segment is supported by the matcher implementation.

## Action / audience / subject / scope examples

```pkl
amends "schema.pkl"

version = 1
issuers { "https://issuer.oathmesh.tech" }
audiences { "https://inventory.internal" }

rules {
  // Subject + action allow
  new {
    name = "storefront-read"
    match {
      sub = "agent://repo/acme/storefront-*"
      act = "inventory.read"
    }
    allow = true
  }

  // Scope-gated write
  new {
    name = "storefront-write-scoped"
    match {
      sub = "agent://repo/acme/storefront-*"
      act = "inventory.write"
      scope {
        "inventory.write"
      }
    }
    allow = true
  }

  // Required default deny
  new {
    name = "default"
    allow = false
  }
}
```

Audience and issuer are enforced before policy evaluation in verification (trusted issuer list + exact audience check). Keep `issuers` and `audiences` aligned with verifier config.

## Troubleshooting policy denials

If you see `policy_denied`:

1. Check denied rule name in error/audit event.
2. Confirm rule order (an early deny can shadow later allow).
3. Verify exact `act`, `env`, `tenant`.
4. Verify `scope` contains all required values.
5. Verify wildcard pattern shape (single `*` glob semantics).
6. Confirm source fields (`src.type`, `src.repo`, `src.workflow`) are present and match expected token claims.

If you see `audience_mismatch` or `issuer_untrusted`, fix verifier configuration first; policy rules are not evaluated until earlier verification steps pass.
