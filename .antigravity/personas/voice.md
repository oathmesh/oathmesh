---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "OathMesh voice — tone, preferred terminology, words to avoid in all communication"
---

# OathMesh Voice

This file defines how OathMesh communicates — in documentation, error messages, marketing copy, code comments, and AI conversation. Consistency builds trust. Inconsistency breaks it.

## Tone Principles

1. **Clear over clever** — say exactly what you mean. No puns, no wordplay, no ambiguity.
2. **Direct over diplomatic** — "This is wrong because X" not "Perhaps we might consider that this could potentially be suboptimal."
3. **Confident over cautious** — OathMesh knows what it is. State it. No hedging on the core value prop.
4. **Technical over casual** — the audience is engineers. Respect their intelligence. Don't oversimplify.
5. **Concise over comprehensive** — say it in fewer words. If a sentence doesn't add value, delete it.

## Preferred Terminology

Use these terms consistently. See `context/glossary.md` for the complete glossary.

| Use | Do Not Use |
|---|---|
| Oath Token | auth token, access token, mesh token |
| Caller | client, user (for machines), requestor |
| Receiver | server, resource server, API server |
| Verified Caller Context | auth context, identity object |
| Issuer | token server, auth server, identity provider (unless external) |
| Policy file | rules file, config file (for policy), access list |
| Mint (a token) | create, generate, issue (in code — "issue" is OK in prose) |
| Verify (a token) | validate, check, authenticate (for token verification) |
| Short-lived signed identity | temporary credentials, ephemeral tokens (too generic) |
| Machine call | API call (only when specifically about APIs), service request |

## Words to Avoid

| Word | Why | Use Instead |
|---|---|---|
| "Simple" | Subjective, often wrong from the reader's perspective | "Focused", "narrow", "purpose-built" |
| "Easy" | Dismisses the reader's potential difficulty | "Straightforward", show-don't-tell with a code sample |
| "Just" | Minimizes complexity ("just add middleware") | Remove the word entirely |
| "Obviously" | Condescending | State the fact without the qualifier |
| "Platform" (for v1) | Overpromises — OathMesh v1 is a protocol + toolkit | "Protocol", "toolkit" |
| "Zero trust" | Overloaded marketing term | "Default-deny policy", "explicit verification" |
| "Seamless" | Meaningless buzzword | Describe what actually happens |
| "Powerful" | Vague | Describe the specific capability |

## Error Message Voice

Error messages are documentation. They must follow this structure:

```
{error_code}: {what went wrong, specifically}

  Expected: {what was expected}
  Received: {what was actually received}

  {One sentence on how to fix it.}

  See: {documentation URL}
```

Example:
```
audience_mismatch: token was minted for https://billing.internal but received by https://inventory.internal

  Expected audience: https://inventory.internal
  Token audience:    https://billing.internal

  Ensure the caller requests a token with the correct audience for this receiver.

  See: https://docs.oathmesh.dev/errors/audience_mismatch
```

## Code Comment Voice

- Explain WHY, not WHAT (the code shows what)
- Use imperative mood for function docs: "Verifies the token signature" not "This function verifies..."
- No humor in code comments
- Reference spec sections when implementing protocol behavior: `// Per oathmesh spec §12: verify audience exact match`

## Documentation Voice

- Guides: second person ("you"), imperative mode ("Run the server", "Add the middleware")
- Reference: third person ("The verifier checks...", "The issuer returns...")
- No first person plural ("we") except in ADRs ("We decided to...")
- Active voice always: "The receiver verifies the token" not "The token is verified by the receiver"
