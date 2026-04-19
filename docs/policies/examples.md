# Policy Examples

Progressive, production-oriented examples based on current matcher behavior.

## 1) Basic allowlist

```pkl
amends "schema.pkl"

version = 1
issuers { "https://issuer.oathmesh.tech" }
audiences { "https://inventory.internal" }

rules {
  new {
    name = "allow-storefront-read"
    match {
      sub = "agent://repo/acme/storefront-bot"
      act = "inventory.read"
    }
    allow = true
  }

  new {
    name = "default"
    allow = false
  }
}
```

## 2) Environment-scoped access

```pkl
amends "schema.pkl"

version = 1
issuers { "https://issuer.oathmesh.tech" }
audiences { "https://inventory.internal" }

rules {
  new {
    name = "prod-read-only"
    match {
      sub = "agent://repo/acme/*"
      act = "inventory.read"
      env = "prod"
    }
    allow = true
  }

  new {
    name = "default"
    allow = false
  }
}
```

## 3) Action + source scoped writes

```pkl
amends "schema.pkl"

version = 1
issuers { "https://issuer.oathmesh.tech" }
audiences { "https://inventory.internal" }

rules {
  new {
    name = "deploy-write-only"
    match {
      sub = "agent://repo/acme/deploy-bot"
      act = "inventory.write"
      src {
        type = "github_actions"
        repo = "acme/storefront"
        workflow = "deploy.yml"
      }
      scope {
        "inventory.write"
      }
    }
    allow = true
  }

  new {
    name = "default"
    allow = false
  }
}
```

## 4) Wildcard usage (supported)

```pkl
amends "schema.pkl"

version = 1
issuers { "https://issuer.oathmesh.tech" }
audiences { "https://inventory.internal" }

rules {
  // prefix*
  new {
    name = "prefix-subject"
    match {
      sub = "agent://repo/acme/*"
      act = "inventory.read"
    }
    allow = true
  }

  // *suffix
  new {
    name = "workflow-suffix"
    match {
      act = "inventory.write"
      src {
        workflow = "*.yml"
      }
    }
    allow = true
  }

  // pre*suf
  new {
    name = "middle-wildcard-action"
    match {
      act = "inventory.*"
    }
    allow = true
  }

  new {
    name = "default"
    allow = false
  }
}
```

## Anti-patterns

- **Allow-all first rule**:
  - `match {}` + `allow = true` at top makes later rules unreachable.
- **Relying on rule order accidentally**:
  - broad allow before specific deny can bypass intended block.
- **Assuming regex syntax**:
  - matcher supports only simple `*` glob forms, not regex.
- **Missing required default rule**:
  - policy validation fails unless last rule is `name = "default", allow = false`.

## Safe defaults

- Start narrow (`sub` + `act` + `env`/`tenant` where applicable).
- Require `scope` for mutate actions.
- Add `src` constraints for CI/CD callers.
- Keep broad denies before broad allows when both may match.
- Keep policy review tied to deployment changes.

See also: [Policy Overview](overview.md) and [Policy Migration](migration.md).
