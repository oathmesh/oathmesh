---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "How to resolve contradictions between rules, agents, skills, and decisions"
---

# OathMesh Conflict Resolution Protocol

When two or more sources of guidance contradict each other, follow this protocol. Do not resolve conflicts silently — always document the conflict and the resolution path taken.

## Priority Hierarchy

Resolve conflicts using this fixed priority order:

| Priority | Category | Governing Files | Overrides |
|---|---|---|---|
| **1 (highest)** | Security | `security/redlines.md`, `rules/security-redlines.md` | Everything below |
| **2** | Architecture | `decisions/ADR-*.md`, `rules/core.md` (architectural constraints) | Performance, Style |
| **3** | Performance | `rules/core.md` (performance budgets) | Style |
| **4 (lowest)** | Style | `rules/coding-standards.md`, `personas/voice.md` | Nothing |

## Resolution Process

### Step 1: Identify the conflict

State clearly:
- What are the two contradicting guidelines?
- Which files contain them?
- What specific section/rule in each file?

### Step 2: Classify each guideline

Assign each conflicting guideline to one of the four priority categories (Security, Architecture, Performance, Style).

### Step 3: Apply the hierarchy

The higher-priority category wins. If both guidelines are in the same category, proceed to Step 4.

### Step 4: Same-category conflict

If both guidelines are in the same priority category:

1. **Check dates** — the more recently modified file takes precedence (check `last_modified` in metadata)
2. **Check specificity** — a rule targeting a specific module overrides a general rule
3. **Check ADRs** — if an ADR addresses this specific conflict, the ADR decision governs
4. **If still ambiguous** — escalate to the human for a decision, and record the resolution as a new ADR or rule update

### Step 5: Document the resolution

After resolving a conflict:

1. Add a note to both source files referencing the resolution
2. If the conflict is systemic (likely to recur), create a new rule in the appropriate `rules/` file
3. If the conflict reveals a gap, propose a new ADR in `feedback/proposals/`

## Agent Conflict Resolution

When two specialized agents (see `agents/`) disagree:

| Scenario | Resolution |
|---|---|
| Security Agent vs. any other agent | Security Agent wins. Always. |
| Architect Agent vs. Test Agent | Architect Agent wins on structural decisions; Test Agent wins on test strategy |
| Architect Agent vs. Docs Agent | Architect Agent wins on technical accuracy; Docs Agent wins on presentation |
| Test Agent vs. Docs Agent | Each governs their own domain — no overlap expected |
| Any tie | Lead Agent arbitrates using this priority hierarchy |

The Lead Agent (`agents/lead-agent.md`) is the final arbiter for all inter-agent conflicts.

## Rule Deprecation

When a rule is superseded:

1. Move the old rule to the `deprecated/` subfolder within its category (e.g., `rules/deprecated/`)
2. Add a deprecation header to the moved file:
   ```yaml
   deprecated: "2026-XX-XX"
   superseded_by: "rules/new-rule-file.md"
   reason: "Brief explanation of why this was superseded"
   ```
3. Update all files that referenced the old rule to point to the new one
4. Log the change in `CHANGELOG.md`
5. Never delete deprecated files — they are historical record

## Escalation Triggers

Escalate to the human operator immediately if:

- Two security rules contradict each other
- A proposed change would violate an active ADR
- The conflict involves an off-limits zone (see `context/current-phase.md`)
- No resolution can be found after applying all steps above
- The conflict involves irreversible actions listed in `security/redlines.md`
