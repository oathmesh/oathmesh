---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Memory expiry and staleness policy — when entries go stale and how to handle them"
---

# Memory Expiry Policy

Every memory entry in `.antigravity/memory/` has a lifecycle. This policy ensures memories stay current and don't silently drift from reality.

## Expiry Rules

### File-Level Metadata

Every memory file must include in its YAML frontmatter:

```yaml
created: "YYYY-MM-DD"        # when the file was first created
last_modified: "YYYY-MM-DD"  # when the file was last meaningfully updated
review_by: "YYYY-MM-DD"      # when the file should be reviewed for accuracy
```

### Default Review Cycles

| File | Default Review Cycle | Rationale |
|---|---|---|
| `memory/project.md` | 90 days | Architecture changes slowly; quarterly review is sufficient |
| `memory/entities.md` | 90 days | Entity list changes infrequently |
| `context/current-phase.md` | 14 days | Phase transitions happen frequently |
| `context/glossary.md` | 180 days | Terminology is stable once defined |
| Session memory files | 90 days | Relevance decays; move to stale after review cycle |
| `decisions/ADR-*.md` | Never expires | ADRs are permanent historical record |

### Staleness Triggers

A memory entry becomes stale when ANY of these conditions are true:

1. **Time-based**: `review_by` date has passed without a review
2. **Event-based**: The source code contradicts what the memory says
3. **Decision-based**: An ADR supersedes information in the memory
4. **External-based**: An external dependency or system referenced has changed significantly

### Handling Stale Entries

When a memory entry is identified as stale:

1. **Do not delete it** — move it to `memory/stale/`
2. **Rename** the file with a staleness timestamp: `original-name.stale-YYYY-MM-DD.md`
3. **Add** a staleness header:
   ```yaml
   stale_since: "YYYY-MM-DD"
   reason: "Brief explanation of why this is stale"
   recovery_action: "What needs to happen to make this current"
   ```
4. **Create** a fresh memory entry if the information is still needed
5. **Report** to the human during the next session that stale entries exist

### Memory Update Triggers

The AI should proactively update memory files when:

| Trigger | Action |
|---|---|
| A new ADR is created | Update `memory/project.md` § Key Design Decisions |
| A new module is added to the codebase | Update `memory/project.md` § Module Map and `memory/entities.md` |
| A new external dependency is introduced | Update `memory/entities.md` § External Systems |
| Technical debt is created or resolved | Update `memory/project.md` § Known Technical Debt |
| A blocker is discovered or resolved | Update `context/current-phase.md` § Known Blockers |
| Phase transition occurs | Update `context/current-phase.md` § Active Phase |
| A new team member joins | Update `memory/entities.md` § People & Roles |
| A competitor ships something relevant | Update `memory/entities.md` § Competitive Landscape |

### Corruption Detection

A memory entry is considered corrupted when:

- It references files or components that don't exist
- It contradicts an active ADR
- It describes architecture that doesn't match the source code
- Its YAML frontmatter is malformed

See `mcp/recovery.md` § Failure Class 3 for the corruption recovery protocol.

## Stale Review Protocol

During the weekly review (see `feedback/weekly-review-prompt.md`), check:

1. Are any `review_by` dates in the past?
2. Are any entries in `memory/stale/` older than 90 days?
3. Does `memory/project.md` match the current state of the codebase?
4. Has `context/current-phase.md` been updated within the last 14 days?

If any check fails, flag it to the human with the specific file and issue.
