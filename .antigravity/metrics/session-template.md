---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Per-session metrics template"
---

# Session Metrics Template

Record at the end of each session. Save to `metrics/sessions/session-YYYY-MM-DD-NN.md`.

```markdown
---
date: "YYYY-MM-DD"
session: NN
---

# Session Metrics — YYYY-MM-DD #NN

## Productivity
| Metric | Value |
|---|---|
| Tasks attempted | N |
| Tasks completed | N |
| Lines of code written | N |
| Lines of code deleted | N |
| Files created | N |
| Files modified | N |
| Commits made | N |

## Quality
| Metric | Value |
|---|---|
| Tests written | N |
| Tests passing | N/N |
| Lint warnings | N |
| Type errors | N |
| Human corrections | N |
| Reversions (undo/revert) | N |

## AI Performance
| Metric | Value |
|---|---|
| Skills triggered | list |
| Agents activated | list |
| Persona used | name |
| Redline checks | N |
| Confidence level | High/Medium/Low |
| Context resets | N |
| MCP server issues | N |

## Phase Progress
| Metric | Value |
|---|---|
| Active phase | Phase N |
| Sprint items completed | N/N |
| Blockers resolved | N |
| New blockers found | N |
| ADRs created | N |
| Tech debt items created | N |
| Tech debt items resolved | N |
```
