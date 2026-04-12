---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Safe, scoped, high-signal first task for orientation"
review_by: "2026-05-05"
---

# First Task: Create the Token Claim JSON Schema

## Why This Task

This task is ideal for orientation because it:
- Touches the core of OathMesh (the token format) without modifying any runtime code
- Requires reading and understanding the specification
- Produces a concrete, testable artifact
- Is low-risk (a JSON Schema file, not production code)
- Exercises the key skills: data modeling, protocol understanding, glossary awareness

## What to Build

Create a JSON Schema file at `/spec/oath-token-claims.schema.json` that formally defines:

1. **Required claims** (`iss`, `sub`, `aud`, `act`, `iat`, `exp`, `jti`) with types, formats, and constraints
2. **Optional claims** (`scope`, `reason`, `src`, `delegated_by`, `env`, `tenant`, `rqh`) with types and constraints
3. **Source object schemas** for each profile (CI, Kubernetes, Agent)
4. **Validation rules** (e.g., `exp > iat`, `sub` matches a URI scheme pattern, `act` is lowercase dot-separated)

## Reference Materials

- `oathmesh.txt` sections 9.3–9.5 (claim definitions)
- `.antigravity/skills/data-modeling.md` (claim schemas and examples)
- `.antigravity/context/glossary.md` (claim term definitions)
- `.antigravity/decisions/ADR-001-token-format.md` (frozen format decisions)

## Steps

1. Read the reference materials above
2. Create `/spec/` directory if it doesn't exist
3. Write the JSON Schema following JSON Schema Draft 2020-12
4. Validate the schema against the example tokens in `oathmesh.txt` sections 9.3–9.4
5. Add a brief README at `/spec/README.md` explaining what the schema covers

## Success Criteria

- [ ] JSON Schema is valid (parseable by any JSON Schema validator)
- [ ] All seven required claims are defined with correct types
- [ ] All optional claims are defined
- [ ] Source object schemas exist for CI, Kubernetes, and Agent profiles
- [ ] The schema validates the example tokens from `oathmesh.txt`
- [ ] The schema rejects tokens missing required claims
- [ ] Uses canonical terminology from `context/glossary.md`

## What This Task Teaches

After completing this task, you will understand:
- The Oath Token claim structure deeply
- How the specification maps to formal schemas
- The relationship between `act` and `scope`
- The subject URI scheme conventions
- How OathMesh profiles extend the base claims
