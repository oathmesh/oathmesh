---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Structured session log format for tracking what worked and what failed"
---

# Session Log Template

Use this template at the end of each session to record outcomes. Save to `feedback/logs/session-YYYY-MM-DD-NN.md`.

```markdown
# Session Log — YYYY-MM-DD #NN

## Session Metrics

| Metric | Value |
|---|---|
| Tasks attempted | N |
| Tasks completed successfully | N |
| Tasks failed or abandoned | N |
| Human corrections applied | N |
| Errors or reversions | N |
| Skills triggered | list |
| Agents activated | list |
| Persona used | name |
| Redline checks performed | N |
| Confidence level | High/Medium/Low |

## What Worked Well

- Describe successful approaches, patterns, or tools that performed as expected

## What Failed or Was Difficult

- Describe failures, unexpected issues, or approaches that didn't work
- Include root cause if known

## Human Corrections

- List any corrections the human made to AI output
- Analyze why the correction was needed
- Propose rule/skill updates if the correction reveals a systemic gap

## Improvement Proposals

If this session revealed a gap in rules, skills, or memory:

- File: `feedback/proposals/proposal-YYYY-MM-DD-NN-brief-title.md`
- Description: what should change and why
- Priority: High/Medium/Low
- Status: Proposed (awaiting human approval)

## Notes for Future Sessions

- Any context that would be valuable for the next session
- Unfinished work and its current state
```
