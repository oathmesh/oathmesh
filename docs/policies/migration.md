# Policy Migration Guide

Move from permissive behavior to fail-closed policy enforcement safely.

## Starting points

Common permissive setups:

- Verifier runs with `PolicyEvaluator = nil` (authenticated tokens allowed).
- Policy includes broad allow rules with little scoping.

Target state:

- Policy evaluator enabled.
- Least-privilege allow rules.
- Explicit default deny as final rule.

## Recommended rollout

### 1) Stage in non-production

1. Enable `PolicyEvaluator` in staging.
2. Start with a compatibility policy:
   - existing required `issuers`/`audiences`
   - temporary broad allow rule
   - required default deny final rule
3. Confirm policy loads/validates and hot-reload behavior is stable.

### 2) Audit-first tightening

1. Replace broad allow with explicit allow rules per subject/action/use-case.
2. Keep default deny as final safety net.
3. Watch allow/deny audit events for missed legitimate paths.
4. Add missing precise allow rules; avoid re-introducing broad allow-all.

### 3) Canary in production

1. Roll strict policy to a subset of traffic/services.
2. Track:
   - `policy_denied`
   - `audience_mismatch`
   - `issuer_untrusted`
3. Fix configuration/policy gaps quickly, then expand rollout.

### 4) Full fail-closed

- Remove temporary compatibility rules.
- Keep only explicit allows + default deny.
- Require change review for any wildcard/broad rule additions.

## Migration checklist

- [ ] Policy ends with `default` deny rule.
- [ ] No top-level allow-all rule in production policy.
- [ ] Write actions require both `act` and `scope`.
- [ ] CI/CD writes restricted with `src.type`, `src.repo`, and `src.workflow`.
- [ ] Verifier audience and trusted issuer config matches deployed policy metadata.
- [ ] Deny alerts and audit dashboards are monitored during rollout.

## Example transition

From permissive:

```pkl
rules {
  new {
    name = "compat-allow-all"
    match {}
    allow = true
  }
  new {
    name = "default"
    allow = false
  }
}
```

To fail-closed:

```pkl
rules {
  new {
    name = "storefront-read"
    match {
      sub = "agent://repo/acme/storefront-*"
      act = "inventory.read"
      env = "prod"
    }
    allow = true
  }
  new {
    name = "deploy-write"
    match {
      sub = "agent://repo/acme/deploy-bot"
      act = "inventory.write"
      scope {
        "inventory.write"
      }
      src {
        type = "github_actions"
        repo = "acme/storefront"
        workflow = "deploy.yml"
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
