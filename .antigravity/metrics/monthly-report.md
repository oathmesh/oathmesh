---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Monthly health report format — AI populates and human reviews"
---

# Monthly Health Report Template

Generate at the end of each month. Save to `metrics/monthly/report-YYYY-MM.md`.

```markdown
---
month: "YYYY-MM"
---

# OathMesh Monthly Health Report — YYYY-MM

## Executive Summary
One paragraph: what was accomplished, what is the overall health of the project and the AI environment.

## Session Statistics
| Metric | Value | Trend |
|---|---|---|
| Total sessions | N | ↑↓→ vs last month |
| Tasks completed | N | ↑↓→ |
| Tasks failed/abandoned | N | ↑↓→ |
| Human corrections | N | ↑↓→ (lower is better) |
| Reversions | N | ↑↓→ (lower is better) |

## Code Quality
| Metric | Value | Trend |
|---|---|---|
| Test coverage (overall) | N% | ↑↓→ |
| Lint warnings | N | ↑↓→ (lower is better) |
| Type errors | N | ↑↓→ (zero is target) |
| Open tech debt items | N | ↑↓→ |
| ADRs created this month | N | — |

## Phase Progress
| Phase | Status | % Complete |
|---|---|---|
| Phase 0 — Protocol Freeze | Active/Complete | N% |
| Phase 1 — MVP Build | Not Started/Active/Complete | N% |

## AI Environment Health
| Check | Status | Notes |
|---|---|---|
| Memory freshness | OK/STALE | List stale entries |
| Decision alignment | OK/DRIFT | List misaligned ADRs |
| Rule effectiveness | OK/GAPS | List rule gaps |
| Skill relevance | OK/OUTDATED | List outdated skills |
| MCP server availability | OK/DEGRADED | List unavailable servers |
| Security posture | OK/ISSUES | List issues |

## Drift Detection
Is the AI's output quality improving, stable, or degrading?

Evidence:
- Human correction trend: [increasing/stable/decreasing]
- Reversion trend: [increasing/stable/decreasing]
- Confidence calibration: [well-calibrated/over-confident/under-confident]

## Key Decisions Made
List ADRs created this month and their significance.

## Risks and Concerns
List any emerging risks or concerns about the project or the AI environment.

## Recommendations
Prioritized list of improvements for next month.
```
