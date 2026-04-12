---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "First-time activation guide for the .antigravity/ AI environment"
---

# OathMesh AI Environment — Setup Guide

## Prerequisites

Before activating this environment, confirm the following:

1. **The OathMesh specification exists** at `oathmesh.txt` in the project root. This is the source-of-truth product document.
2. **The AI assistant** (Antigravity or equivalent) is configured to read from this project's workspace root.
3. **Git is initialized** in the project root. The `.antigravity/` folder should be committed and version-controlled.
4. **No `.env` files contain secrets** in the project root. Secrets must never be committed. See `security/secret-policy.md`.

## First-Run Checklist

Complete these steps in order before the AI takes its first action:

### Step 1 — Verify directory integrity

Confirm all required directories exist:

```
.antigravity/
  context/     — must contain: system-prompt.md, glossary.md, current-phase.md
  rules/       — must contain: core.md, coding-standards.md, conflict-resolution.md, security-redlines.md
  skills/      — must contain: auth.md, protocol-transport.md, api-design.md, data-modeling.md, identity-resolution.md
  mcp/         — must contain: servers.json, fallbacks.md, recovery.md
  memory/      — must contain: project.md, entities.md, session-template.md, expiry-policy.md
  decisions/   — must contain: ADR-000-template.md, ADR-001-token-format.md, ADR-002-auth-strategy.md, ADR-003-tech-stack.md
  agents/      — must contain: lead-agent.md, architect-agent.md, security-agent.md, test-agent.md, docs-agent.md
  personas/    — must contain: voice.md, senior-engineer.md, debug-mode.md, review-mode.md, rapid-prototype.md
  security/    — must contain: redlines.md, secret-policy.md
  feedback/    — must contain: session-log-template.md, weekly-review-prompt.md
  workflows/   — must contain: feature.md, bugfix.md, hotfix.md, review.md, refactor.md
  onboarding/  — must contain: sequence.md, first-task.md
  testing/     — must contain: philosophy.md
  metrics/     — must contain: session-template.md, monthly-report.md
  tools/       — must contain: git.md, linting.md, ci-cd.md
```

### Step 2 — Human review of critical files

These files must be reviewed and confirmed by a human before the AI acts:

- [ ] `context/current-phase.md` — Is the active phase correct?
- [ ] `security/redlines.md` — Are all redlines accepted?
- [ ] `security/secret-policy.md` — Are secret handling rules correct for your environment?
- [ ] `rules/core.md` — Are the non-negotiable rules accepted?
- [ ] `decisions/ADR-003-tech-stack.md` — Is the tech stack decision confirmed?

### Step 3 — Configure MCP servers

Review `mcp/servers.json` and update server endpoints for your local environment. At minimum, the filesystem and git servers must be reachable.

### Step 4 — Boot the AI

Instruct the AI to read `README.md` and follow the boot sequence. The AI should:

1. Read `README.md` fully
2. Load `context/system-prompt.md`
3. Load `context/current-phase.md`
4. Load `context/glossary.md`
5. Load `rules/core.md`
6. Load `rules/conflict-resolution.md`
7. Load `memory/project.md`
8. Load `memory/entities.md`
9. Scan `decisions/` for all active ADRs
10. Confirm boot is complete

### Step 5 — Run health check

Follow `health-check.md` to verify the environment is fully operational.

### Step 6 — Begin work

Start with `onboarding/first-task.md` if this is the first session. Otherwise, check `context/current-phase.md` for active priorities.

## Post-Setup Maintenance

- After every significant session, update `memory/project.md` if architecture changed.
- After every major decision, create a new ADR in `decisions/`.
- Run `health-check.md` weekly.
- Run `feedback/weekly-review-prompt.md` weekly.
- Review `memory/stale/` monthly for entries needing human validation.
