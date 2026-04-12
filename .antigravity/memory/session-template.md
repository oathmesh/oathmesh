---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Template for session-end memory writes — captures what happened and what changed"
---

# Session Memory Template

Copy this template at the end of each significant session. Save as `memory/session-YYYY-MM-DD-NN.md` where NN is the session number for that day (01, 02, etc.).

---

```markdown
---
session_date: "YYYY-MM-DD"
session_number: NN
duration_estimate: "X hours"
review_by: "YYYY-MM-DD"  # 90 days from session date
---

# Session Summary — YYYY-MM-DD #NN

## What Was Done

- [ ] Briefly list each task completed
- [ ] Use action verbs: implemented, fixed, designed, refactored, documented

## Decisions Made

| Decision | Rationale | ADR Created? |
|---|---|---|
| Describe the decision | Why this was chosen | Yes/No — if yes, link to ADR |

## Files Changed

| File | Change Type | Description |
|---|---|---|
| `path/to/file` | Created / Modified / Deleted | Brief description of change |

## Architecture Changes

Describe any changes to component boundaries, data flow, or public API surface.
If none: "No architecture changes."

## New Technical Debt

| ID | Description | Priority |
|---|---|---|
| TD-NNN | What was deferred and why | High/Medium/Low |

## Blockers Encountered

| Blocker | Resolution | Still Blocking? |
|---|---|---|
| What blocked progress | How it was resolved or worked around | Yes/No |

## Memory Updates Required

List any updates needed to persistent memory files:

- [ ] `memory/project.md` — update if architecture changed
- [ ] `memory/entities.md` — update if new components/systems added
- [ ] `context/current-phase.md` — update if priorities changed
- [ ] `CHANGELOG.md` — update if .antigravity/ config changed

## Next Session Recommendations

What should be done next, in priority order:

1. First priority task
2. Second priority task
3. Third priority task

## AI Self-Assessment

- Tasks completed successfully: N/N
- Errors or reversions: describe any
- Human corrections applied: describe any
- Skills triggered: list which skills from skills/ were relevant
- Confidence level in session output: High/Medium/Low
```
