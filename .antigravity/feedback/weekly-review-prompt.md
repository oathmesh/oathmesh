---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Weekly self-review prompt — run against memory to surface outdated assumptions"
---

# Weekly Review Prompt

Run this review weekly (or after every 5 significant sessions). The AI should execute each check and report findings to the human.

## Memory Freshness Review

1. **Check `memory/project.md`**: Does the Module Map match the actual files in the repository? List any discrepancies.

2. **Check `memory/entities.md`**: Are all listed components still current? Have any new components been added that aren't listed?

3. **Check `context/current-phase.md`**: Is the active phase still correct? Have any blockers been resolved but not updated?

4. **Check `memory/stale/`**: Are there any entries older than 90 days? List them for human review.

5. **Check all `review_by` dates**: List any files where `review_by` has passed.

## Decision Alignment Review

6. **Scan `decisions/`**: Are all accepted ADRs still aligned with the current codebase? Flag any ADR whose decision no longer matches reality.

7. **Check for undocumented decisions**: Were any architectural decisions made in recent sessions without a corresponding ADR? List them.

## Rule Effectiveness Review

8. **Check `rules/coding-standards.md`**: Are the coding standards still appropriate for the current codebase size and complexity?

9. **Check `rules/security-redlines.md`**: Were any redline checks bypassed or ignored in recent sessions? If so, was the bypass justified?

10. **Check `rules/conflict-resolution.md`**: Were there any unresolved conflicts in recent sessions?

## Skill Relevance Review

11. **Check `skills/`**: Are all skill trigger conditions still accurate? Has the codebase evolved in a way that makes a skill obsolete?

12. **Check for skill gaps**: Were there situations in recent sessions where no skill was applicable but one should have existed?

## Self-Assessment Questions

13. **Pattern detection**: Are there repeated human corrections across sessions? If so, what rule or skill change would prevent them?

14. **Drift detection**: Is the AI's output quality improving, stable, or degrading over time? (Compare recent session logs.)

15. **Confidence calibration**: Were there cases where the AI was confident but wrong? Were there cases where the AI was uncertain but the answer was obvious in hindsight?

## Report Format

```
Weekly Review — YYYY-MM-DD
Reviewer: AI

Memory Freshness:     [OK | N issues found]
Decision Alignment:   [OK | N misalignments found]  
Rule Effectiveness:   [OK | N gaps found]
Skill Relevance:      [OK | N updates needed]
Self-Assessment:      [Improving | Stable | Degrading]

Issues Found:
1. [issue description] — [recommended action]
2. [issue description] — [recommended action]

Proposals Generated: N (see feedback/proposals/)
```

## Improvement Trigger

If this review surfaces any of the following, create a proposal in `feedback/proposals/`:

- A rule that was consistently violated or bypassed
- A skill that was triggered but produced incorrect guidance
- A memory entry that no longer matches reality
- A missing ADR for a decision made without one
- A new term that should be in the glossary
- A new threat that should be in the threat model

Proposals require human approval before being applied. See `feedback/proposals/.gitkeep` for the proposals directory.
