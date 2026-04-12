---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Track all changes to the .antigravity/ configuration system"
---

# OathMesh AI Environment — Changelog

All notable changes to this AI configuration environment are documented here. Every file in `.antigravity/` carries its own version number. This changelog tracks system-wide changes.

Format: `[version] — YYYY-MM-DD — author — summary`

---

## [1.0] — 2026-04-05 — Founder — Initial scaffold

### Created
- Full `.antigravity/` directory structure (19 directories, 50+ files)
- Boot sequence defined in `README.md`
- System prompt, glossary, and current phase in `context/`
- Core rules, coding standards, conflict resolution, security redlines in `rules/`
- Five domain skills: auth, protocol-transport, api-design, data-modeling, identity-resolution in `skills/`
- MCP server configuration with fallback and recovery plans in `mcp/`
- Project memory, entity memory, session templates, and expiry policy in `memory/`
- Three seed ADRs: ADR-001 (token format), ADR-002 (auth strategy), ADR-003 (tech stack) in `decisions/`
- Five specialized agents: lead, architect, security, test, docs in `agents/`
- Five communication personas: voice, senior-engineer, debug-mode, review-mode, rapid-prototype in `personas/`
- Security redlines and secret policy in `security/`
- Session logging and weekly review prompts in `feedback/`
- Five workflows: feature, bugfix, hotfix, review, refactor in `workflows/`
- Onboarding sequence and first-task guide in `onboarding/`
- Testing philosophy in `testing/`
- Session metrics and monthly report templates in `metrics/`
- Git, linting, and CI/CD toolchain rules in `tools/`

### Design Decisions
- Rule priority hierarchy: Security > Architecture > Performance > Style
- Memory expiry policy: 90-day default review cycle
- ADR naming: `ADR-NNN-kebab-case-title.md`
- Agent activation: condition-based, not always-on
- Persona selection: context-driven with explicit overrides

---

## Pending

_No pending changes._
