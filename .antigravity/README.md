---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "AI orientation document — read this first at every session start"
---

# OathMesh AI Environment

This folder is the persistent AI brain for the OathMesh project. It is read at the start of every session and updated as the project evolves. If you are an AI assistant, read this file completely before taking any action.

## What Is OathMesh

OathMesh is a micro-protocol and developer product that replaces shared machine secrets (API keys, static tokens) with short-lived, signed JWT-based call identity. It is built for agents, CI/CD jobs, internal tools, automation bots, and service-to-service calls.

**One-line positioning:** OathMesh gives every machine call a short-lived signed identity.

**Core doctrine:** OathMesh authenticates the caller. The receiver authorizes the request.

## Boot Sequence

At the start of every session, load context in this order:

1. **This file** — orientation and system map
2. **`context/system-prompt.md`** — active system prompt and identity
3. **`context/current-phase.md`** — what phase we are in, active priorities, blockers
4. **`context/glossary.md`** — OathMesh domain terms (use these consistently)
5. **`rules/core.md`** — non-negotiable behavioral rules
6. **`rules/conflict-resolution.md`** — how to resolve contradictions
7. **`memory/project.md`** — architecture decisions and current state
8. **`memory/entities.md`** — key components, services, and systems
9. **`decisions/`** — check ADRs before suggesting architectural changes

Do not skip steps. Do not reorder. If any file is missing or corrupt, halt and report which file is unavailable before proceeding.

## Directory Map

```
.antigravity/
├── README.md                          ← You are here. Read first every session.
├── CHANGELOG.md                       ← All changes to this configuration tracked here.
├── setup.md                           ← First-time activation guide.
├── health-check.md                    ← Verification protocol for system integrity.
│
├── context/                           ← Session-start context and active state.
│   ├── system-prompt.md               ← AI identity and behavioral contract.
│   ├── glossary.md                    ← OathMesh terminology — enforced in all output.
│   └── current-phase.md              ← Active sprint, priorities, blockers, off-limits zones.
│
├── rules/                             ← Behavioral and coding rules. Always authoritative.
│   ├── core.md                        ← Non-negotiable rules. Priority hierarchy defined here.
│   ├── coding-standards.md            ← Naming, formatting, language conventions per module.
│   ├── conflict-resolution.md         ← How to resolve contradictions between rules/agents.
│   ├── security-redlines.md           ← Actions that ALWAYS require human approval.
│   └── deprecated/                    ← Old rules kept for history. Never deleted.
│
├── skills/                            ← Domain-specific AI capabilities for OathMesh work.
│   ├── auth.md                        ← JWT/JWS, OIDC, key management, token lifecycle.
│   ├── protocol-transport.md          ← HTTP/gRPC transport bindings, gateway mode.
│   ├── api-design.md                  ← Issuer API, JWKS, metadata endpoints.
│   ├── data-modeling.md               ← Token schemas, policy schemas, audit events.
│   ├── identity-resolution.md         ← Subject URI schemes, caller identity patterns.
│   └── deprecated/                    ← Old skills kept for history. Never deleted.
│
├── mcp/                               ← Model Context Protocol server configuration.
│   ├── servers.json                   ← Server definitions, endpoints, auth, scopes.
│   ├── fallbacks.md                   ← What to do when an MCP server is unavailable.
│   └── recovery.md                    ← Recovery playbook for MCP failures.
│
├── memory/                            ← Persistent project knowledge.
│   ├── project.md                     ← Architecture, core abstractions, tech debt.
│   ├── entities.md                    ← Components, services, people, systems.
│   ├── session-template.md            ← Format for session-end memory writes.
│   ├── expiry-policy.md               ← When memories go stale and how to handle them.
│   └── stale/                         ← Entries awaiting human validation.
│
├── decisions/                         ← Architecture Decision Records.
│   ├── ADR-000-template.md            ← ADR template — use this format for all new ADRs.
│   ├── ADR-001-token-format.md        ← Why JWT/JWS with om+jwt type header.
│   ├── ADR-002-auth-strategy.md       ← Caller authentication and receiver-local authz.
│   └── ADR-003-tech-stack.md          ← Language choices: Go issuer, Node/Python SDKs.
│
├── agents/                            ← Specialized sub-agent definitions.
│   ├── lead-agent.md                  ← Coordinator. Owns session state, arbitrates conflicts.
│   ├── architect-agent.md             ← System design, ADR creation, dependency analysis.
│   ├── security-agent.md              ← Threat modeling, redline enforcement, secret scanning.
│   ├── test-agent.md                  ← Coverage strategy, test generation, quality gates.
│   └── docs-agent.md                  ← Documentation sync with code changes.
│
├── personas/                          ← Communication mode definitions.
│   ├── voice.md                       ← OathMesh tone, terminology, words to avoid.
│   ├── senior-engineer.md             ← Precise, direct, no hand-holding.
│   ├── debug-mode.md                  ← Investigative, systematic, hypothesis-driven.
│   ├── review-mode.md                 ← Critical, constructive, long-term quality focus.
│   └── rapid-prototype.md            ← Fast, explicit about shortcuts taken.
│
├── security/                          ← Security policies and threat awareness.
│   ├── redlines.md                    ← Irreversible actions requiring human approval.
│   └── secret-policy.md              ← What to never log, commit, or expose.
│
├── feedback/                          ← Self-improvement and session tracking.
│   ├── session-log-template.md        ← Structured format for session outcomes.
│   ├── weekly-review-prompt.md        ← AI self-review against stale assumptions.
│   └── proposals/                     ← AI-drafted rule/skill changes awaiting approval.
│
├── workflows/                         ← Step-by-step processes for common work patterns.
│   ├── feature.md                     ← Discovery → Design → Implement → Test → Document.
│   ├── bugfix.md                      ← Reproduce → Isolate → Hypothesize → Fix → Verify.
│   ├── hotfix.md                      ← Fast-path with mandatory post-mortem.
│   ├── review.md                      ← Checklist, ADR alignment, security scan, style gate.
│   └── refactor.md                    ← Scope → Safety net → Incremental execution.
│
├── onboarding/                        ← New developer and fresh-session orientation.
│   ├── sequence.md                    ← Structured onboarding sequence.
│   └── first-task.md                  ← Safe, scoped, high-signal first task.
│
├── testing/                           ← Test philosophy and quality expectations.
│   └── philosophy.md                  ← Boundaries, naming, coverage, mocking strategy.
│
├── metrics/                           ← Session quality and system health tracking.
│   ├── session-template.md            ← Per-session metrics format.
│   └── monthly-report.md             ← Monthly health report format.
│
└── tools/                             ← Toolchain integration rules.
    ├── git.md                         ← Commit conventions, branch naming, hooks.
    ├── linting.md                     ← Formatting configs and enforcement rules.
    └── ci-cd.md                       ← Pipeline awareness and build targets.
```

## How Files Interact

```
Boot: README → context/ → rules/ → memory/ → decisions/

During Work:
  rules/core.md ──────────── governs all output
  rules/coding-standards.md ─ governs all code
  skills/* ───────────────── activated by task type
  agents/* ───────────────── activated by context
  personas/* ─────────────── activated by situation
  decisions/* ────────────── checked before architectural changes
  security/redlines.md ───── checked before irreversible actions
  workflows/* ────────────── followed for structured work

Session End:
  feedback/session-log-template.md ── record outcomes
  memory/project.md ──────────────── update if architecture changed
  memory/entities.md ─────────────── update if new components added
  decisions/ ─────────────────────── write ADR if major decision made
  CHANGELOG.md ───────────────────── log config changes if any
```

## Emergency Procedures

If memory is corrupted: see `mcp/recovery.md`
If MCP server is down: see `mcp/fallbacks.md`
If rules conflict: see `rules/conflict-resolution.md`
If uncertain about an action: see `security/redlines.md` then ask the human
If this is a fresh session with no prior context: see `onboarding/sequence.md`

## Rule Hierarchy

When rules conflict, resolve using this fixed priority:

1. **Security** (`security/redlines.md`, `rules/security-redlines.md`)
2. **Architecture** (`decisions/`, `rules/core.md`)
3. **Performance** (`rules/core.md` § performance budgets)
4. **Style** (`rules/coding-standards.md`, `personas/voice.md`)

Full resolution protocol: `rules/conflict-resolution.md`
