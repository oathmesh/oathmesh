---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "System integrity verification protocol — run weekly or after anomalies"
---

# OathMesh AI Environment — Health Check

Run this protocol weekly, after MCP failures, after config changes, or whenever AI behavior seems inconsistent. Report results to the human operator.

## 1. File Integrity Check

Verify every required file exists and is non-empty:

| Category | Files | Status |
|---|---|---|
| Foundation | `README.md`, `CHANGELOG.md`, `setup.md`, `health-check.md` | ☐ |
| Context | `context/system-prompt.md`, `context/glossary.md`, `context/current-phase.md` | ☐ |
| Rules | `rules/core.md`, `rules/coding-standards.md`, `rules/conflict-resolution.md`, `rules/security-redlines.md` | ☐ |
| Skills | `skills/auth.md`, `skills/protocol-transport.md`, `skills/api-design.md`, `skills/data-modeling.md`, `skills/identity-resolution.md` | ☐ |
| MCP | `mcp/servers.json`, `mcp/fallbacks.md`, `mcp/recovery.md` | ☐ |
| Memory | `memory/project.md`, `memory/entities.md`, `memory/session-template.md`, `memory/expiry-policy.md` | ☐ |
| Decisions | `decisions/ADR-000-template.md`, at least one active ADR | ☐ |
| Agents | `agents/lead-agent.md`, `agents/architect-agent.md`, `agents/security-agent.md`, `agents/test-agent.md`, `agents/docs-agent.md` | ☐ |
| Personas | `personas/voice.md`, `personas/senior-engineer.md`, `personas/debug-mode.md`, `personas/review-mode.md`, `personas/rapid-prototype.md` | ☐ |
| Security | `security/redlines.md`, `security/secret-policy.md` | ☐ |
| Feedback | `feedback/session-log-template.md`, `feedback/weekly-review-prompt.md` | ☐ |
| Workflows | `workflows/feature.md`, `workflows/bugfix.md`, `workflows/hotfix.md`, `workflows/review.md`, `workflows/refactor.md` | ☐ |

## 2. Cross-Reference Integrity

Verify that:

- [ ] `README.md` boot sequence references only files that exist
- [ ] `rules/core.md` references to `rules/conflict-resolution.md` and `security/redlines.md` are valid
- [ ] All ADR files follow the template in `decisions/ADR-000-template.md`
- [ ] Agent files reference valid rules, skills, and personas
- [ ] Workflow files reference valid rules and security checks

## 3. Memory Freshness

- [ ] `memory/project.md` — `last_modified` is within 30 days
- [ ] `memory/entities.md` — `last_modified` is within 30 days
- [ ] `context/current-phase.md` — `last_modified` is within 14 days
- [ ] `memory/stale/` — check if any entries are older than 90 days (escalate to human)

## 4. Rule Consistency

- [ ] `rules/core.md` priority hierarchy matches `README.md` rule hierarchy
- [ ] `rules/security-redlines.md` is consistent with `security/redlines.md`
- [ ] No active rules in `rules/deprecated/` are referenced by non-deprecated files
- [ ] `rules/conflict-resolution.md` covers all four priority levels

## 5. MCP Server Availability

For each server in `mcp/servers.json`:

- [ ] Filesystem server — can read project files
- [ ] Git server — can read commit history
- [ ] Terminal server — can execute commands
- [ ] If any server is unavailable — verify fallback in `mcp/fallbacks.md` is actionable

## 6. Skill Relevance

- [ ] No skills in `skills/` reference deprecated libraries or patterns
- [ ] Skill trigger conditions in each skill file are still accurate
- [ ] No deprecated skills are referenced by active agent definitions

## 7. Security Posture

- [ ] `security/redlines.md` — all redlines are still appropriate
- [ ] `security/secret-policy.md` — no secrets have been committed to the repo
- [ ] No `.env` files exist in version control
- [ ] No hardcoded keys, tokens, or passwords in any `.antigravity/` file

## Report Format

After completing the health check, produce a summary:

```
Health Check — [DATE]
Status: [HEALTHY | DEGRADED | CRITICAL]
Files checked: [N/N present]
Memory freshness: [OK | STALE entries found]
MCP servers: [N/N reachable]
Rule consistency: [OK | CONFLICTS found]
Security posture: [OK | ISSUES found]
Actions required: [list or "none"]
```

If status is DEGRADED or CRITICAL, follow `mcp/recovery.md` for recovery procedures and escalate to the human operator.
