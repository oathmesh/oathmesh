---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "OathMesh domain terminology — use these exact terms in all output"
---

# OathMesh Glossary

Use these terms exactly as defined. Do not invent synonyms. Do not abbreviate unless the abbreviation is listed here. When writing documentation, code comments, error messages, or conversational responses, use these canonical forms.

## Core Protocol Terms

| Term | Definition | Never Use Instead |
|---|---|---|
| **Oath Token** | A short-lived signed JWT assertion (`om+jwt` type) minted by an issuer for a specific caller, audience, and action. | "mesh token", "auth token", "access token", "OathMesh token" |
| **Issuer** | The service that authenticates callers and mints signed Oath Tokens. Exposes JWKS and metadata endpoints. | "token server", "auth server", "identity provider" (unless referring to an external IdP) |
| **Caller** | The agent, workflow, bot, service, or tool client that requests an Oath Token and attaches it to outbound requests. | "client" (too generic), "user" (reserved for humans) |
| **Receiver** | The API, tool server, or gateway that accepts incoming requests and verifies the attached Oath Token. | "server" (too generic), "resource server" |
| **Verifier** | The middleware or proxy logic within a Receiver that validates an Oath Token's signature, claims, and freshness. | "validator", "auth middleware" |
| **Verified Caller Context** | The normalized identity object exposed to application code after successful token verification. Contains principal, action, scope, source, and token_id. | "auth context", "identity object", "user context" |
| **Policy Engine** | The component that evaluates whether a verified request is allowed, based on a YAML policy file. Default-deny. | "authorization engine", "rules engine" |

## Token Claims

| Claim | Full Name | Description |
|---|---|---|
| `iss` | Issuer | URI of the issuer that signed the token |
| `sub` | Subject | URI identifying the caller (uses OathMesh subject schemes) |
| `aud` | Audience | URI of the intended receiver |
| `act` | Action | The requested operation family (e.g., `inventory.write`) |
| `iat` | Issued At | Unix timestamp when the token was minted |
| `exp` | Expiry | Unix timestamp when the token expires |
| `jti` | Token ID | Unique identifier for replay detection |
| `scope` | Scope | Array of permitted operations (optional, superset of `act`) |
| `src` | Source | Provenance object — where the caller is running (optional) |
| `rqh` | Request Hash | SHA-256 hash of the bound request (optional) |

## Subject URI Schemes

| Scheme | Use Case | Example |
|---|---|---|
| `svc://` | Service-to-service calls | `svc://payments/billing-sync` |
| `agent://` | Repository-based agents or bots | `agent://repo/acme/deploy-bot` |
| `job://` | Scheduled or CI/CD jobs | `job://github-actions/acme/storefront/deploy` |
| `tool://` | Tool clients (MCP, LLM tools) | `tool://mcp/code-executor` |
| `user://` | Human delegation identity | `user://mustafa` |

## Protocol Profiles

| Profile | Use Case | Key Additions |
|---|---|---|
| **Core Profile** | Basic internal API calls | Signed token, required claims, HTTP transport |
| **CI Profile** | GitHub Actions, GitLab CI, Jenkins | `src.repo`, `src.workflow`, `src.run_id`, `src.sha` |
| **Agent Profile** | AI runtimes, tool execution | `src.session_id`, `src.tool_client`, `delegated_by` |
| **Gateway Profile** | Proxy-based deployments | Trust header normalization, audit forwarding |
| **Kubernetes Profile** | Cluster-native workloads | `src.cluster`, `src.namespace`, `src.pod`, `src.service_account` |

## Transport

| Term | Definition |
|---|---|
| **Oathmesh scheme** | The canonical HTTP authorization header: `Authorization: Oathmesh <token>` |
| **Bearer compatibility** | Fallback mode using `Authorization: Bearer <token>` — only when necessary |
| **Metadata endpoint** | `/.well-known/oathmesh-issuer` — issuer discovery and configuration |
| **JWKS endpoint** | `/.well-known/jwks.json` — public keys for token verification |

## Security Terms

| Term | Definition |
|---|---|
| **TTL** | Time-to-live for an Oath Token. Default: 120 seconds. Maximum: 300 seconds. |
| **Replay cache** | A short-lived cache of `jti` values used to detect replayed tokens on write operations. |
| **Request binding** | Optional mechanism where the token includes a hash (`rqh`) of the HTTP request, binding the token to a specific request. |
| **Key rotation** | The process of replacing the issuer's signing key pair with overlapping public key publication for graceful transition. |
| **Default deny** | The policy engine denies all requests unless an explicit allow rule matches. |

## Product Terms

| Term | Definition |
|---|---|
| **Golden path** | The primary MVP demo: GitHub Actions → Oathmesh Issuer → Internal API |
| **Audit event** | A structured JSON log entry recording a verification or authorization outcome (allow/deny, principal, action, token_id). |
| **Managed issuer** | The hosted, paid version of the issuer service — the first commercial product. |
| **Policy file** | A YAML file defining allow/deny rules evaluated by the policy engine. |

## Words to Avoid

Do not use these terms in OathMesh documentation, code, or conversations:

| Avoid | Reason | Use Instead |
|---|---|---|
| "mesh" alone | Implies service mesh, which OathMesh is not | "OathMesh" (full name) or "protocol" |
| "trust fabric" | Overly abstract | "signed identity" or "verified caller context" |
| "zero trust" as a product claim | Overloaded marketing term | "explicit verification" or "default-deny policy" |
| "universal" | OathMesh is intentionally narrow | "focused" or "purpose-built" |
| "platform" (for v1) | Implies more than what v1 delivers | "protocol" or "toolkit" |
| "user" for machine callers | Confuses human vs. machine identity | "caller" |
| "client" for callers | Too generic | "caller" |
| "secret" for Oath Tokens | Tokens are signed assertions, not secrets | "token" or "Oath Token" |
