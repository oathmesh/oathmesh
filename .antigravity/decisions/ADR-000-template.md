---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Template for all Architecture Decision Records — use this format for every new ADR"
---

# ADR-000: Template

**Status:** Template (do not change status on this file)

## Context

Describe the technical or business context that requires a decision. Include:
- What problem or question arose?
- Why does it need to be decided now?
- What constraints exist?
- Reference to relevant OathMesh spec sections, existing ADRs, or external standards.

## Options Considered

### Option A: [Name]

- Description of the approach
- Pros: what it does well
- Cons: what it does poorly or risks

### Option B: [Name]

- Description of the approach
- Pros: what it does well
- Cons: what it does poorly or risks

### Option C: [Name] (if applicable)

- Description of the approach
- Pros: what it does well
- Cons: what it does poorly or risks

## Decision

State the decision clearly. Use active voice: "We will use X because Y."

## Consequences

### Positive
- What improves as a result of this decision

### Negative
- What trade-offs or limitations this decision introduces

### Risks
- What could go wrong and how to mitigate it

## References

- Links to OathMesh spec sections, RFCs, external documentation
- Links to other ADRs this depends on or conflicts with

---

**Naming convention:** `ADR-NNN-kebab-case-title.md`

**Status values:**
- `Proposed` — under discussion, not yet decided
- `Accepted` — decided and in effect
- `Deprecated` — superseded by a newer ADR (add `superseded_by` field)
- `Template` — this file only

**Rules:**
- Check existing ADRs before proposing a new one (avoid duplicates)
- Deprecated ADRs are never deleted
- Every ADR must reference the OathMesh spec section(s) it relates to
- ADR amendments are new ADRs that reference and supersede the original
