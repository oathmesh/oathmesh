---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
skill: "identity-resolution"
triggers:
  - "constructing subject URIs from bootstrap claims"
  - "designing caller identity formats"
  - "implementing subject matching in policy evaluation"
  - "mapping external identity claims to OathMesh subjects"
  - "working on the delegated_by claim"
dependencies:
  - "context/glossary.md (for subject URI scheme definitions)"
  - "skills/auth.md (for bootstrap identity flows)"
---

# Skill: Identity Resolution

This skill covers how OathMesh constructs, formats, parses, and matches caller identities — the `sub` claim and the subject URI scheme system.

## Subject URI Schemes

OathMesh standardizes caller identity into URI-formatted strings. This ensures policies are clean, matchable, and human-readable.

### Scheme Definitions

| Scheme | Pattern | Use Case | Examples |
|---|---|---|---|
| `svc://` | `svc://{namespace}/{service-name}` | Service-to-service calls | `svc://payments/billing-sync`, `svc://default/api-gateway` |
| `agent://` | `agent://repo/{org}/{name}` | Repository-based agents/bots | `agent://repo/acme/deploy-bot`, `agent://repo/acme/monitor-agent` |
| `job://` | `job://{provider}/{org}/{repo}/{workflow}` | CI/CD jobs | `job://github-actions/acme/storefront/deploy` |
| `tool://` | `tool://{runtime}/{client-name}` | Tool clients (MCP, LLM) | `tool://mcp/code-executor`, `tool://langchain/research-agent` |
| `user://` | `user://{identifier}` | Human delegation identity | `user://mustafa`, `user://eng-team` |

### Construction Rules

1. All scheme names are lowercase, followed by `://`
2. Path segments are lowercase, alphanumeric, hyphens, and underscores only
3. No trailing slashes
4. No query parameters or fragments
5. Maximum total length: 256 characters
6. Callers never self-assert their `sub` — the issuer constructs it from verified bootstrap claims

### GitHub Actions → Subject Mapping

When a GitHub Actions workflow authenticates via OIDC, the issuer constructs the subject as:

```
job://github-actions/{owner}/{repo}/{workflow_name}
```

Where:
- `{owner}` is from the GitHub OIDC `repository_owner` claim
- `{repo}` is from the GitHub OIDC `repository` claim (without owner prefix)
- `{workflow_name}` is from the `workflow` claim, without `.yml` extension, lowercased

Example: Repository `acme/storefront`, workflow `deploy.yml`:
```
job://github-actions/acme/storefront/deploy
```

### Kubernetes → Subject Mapping

When a Kubernetes workload authenticates via ServiceAccount token:

```
svc://{namespace}/{service-account-name}
```

Example: namespace `payments`, service account `billing-sync`:
```
svc://payments/billing-sync
```

### Client Credentials → Subject Mapping (Dev Mode)

In dev mode with pre-shared client credentials:

```
{configured sub from client config}
```

The subject is explicitly set in the client configuration. The issuer does not derive it.

## Subject Matching in Policy

Policy rules match against subject URIs using these patterns:

### Exact Match
```yaml
match:
  sub: "agent://repo/acme/deploy-bot"
```

### Glob Matching
```yaml
match:
  sub: "agent://repo/acme/*"           # matches any agent under acme
  sub: "job://github-actions/acme/*/*"  # matches any workflow in any acme repo
  sub: "svc://payments/*"              # matches any service in payments namespace
```

### Matching Rules

- `*` matches any sequence of characters within a single path segment
- `**` is not supported in v1 (reserved for future use)
- Matching is case-sensitive
- The scheme prefix (`svc://`, `agent://`, etc.) must match exactly
- Empty segments are not allowed in patterns

## Delegation

When a machine call is performed on behalf of a human, the `delegated_by` claim records the human identity:

```json
{
  "sub": "agent://repo/acme/deploy-bot",
  "delegated_by": "user://mustafa"
}
```

### Delegation Rules

- `delegated_by` must use the `user://` scheme
- The issuer must verify the delegation claim — it cannot be self-asserted by the caller
- The issuer can verify delegation via:
  - The bootstrap identity includes a delegation claim (e.g., GitHub OIDC `actor` claim)
  - An explicit delegation grant in the issuer's configuration
- Delegation does not grant additional permissions — the receiver's policy evaluates the `sub`, and `delegated_by` is informational for audit

## Identity Lifecycle

1. **Creation**: Subject URI is constructed by the issuer at token mint time
2. **Transport**: Subject is embedded in the `sub` claim of the Oath Token
3. **Verification**: Subject is extracted and included in the Verified Caller Context
4. **Authorization**: Subject is matched against policy rules
5. **Audit**: Subject is logged in the audit event

The subject URI should be stable for a given caller. The same GitHub Actions workflow in the same repo should always produce the same `sub` value, enabling consistent policy matching and audit correlation over time.
