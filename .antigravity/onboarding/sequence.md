---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Structured onboarding sequence for new developers or fresh AI sessions"
---

# Onboarding Sequence

## Who This Is For

- A new developer joining the OathMesh project
- A fresh AI session with no prior context
- Any agent needing to understand OathMesh from scratch

## Sequence (Follow in Order)

### Step 1: Understand What OathMesh Is (5 minutes)

Read these in order:
1. `.antigravity/README.md` — AI environment orientation
2. `.antigravity/context/system-prompt.md` — mission and scope
3. `oathmesh.txt` sections 1–4 (lines 1–163) — executive summary, positioning, problem, product definition

After this step you should be able to answer:
- What does OathMesh replace?
- What is the core doctrine?
- What is OathMesh NOT?

### Step 2: Understand the Protocol (10 minutes)

Read:
1. `.antigravity/context/glossary.md` — terminology
2. `oathmesh.txt` sections 9–12 (lines 311–514) — token spec, transport, metadata, verification
3. `.antigravity/decisions/ADR-001-token-format.md` — why JWT/JWS

After this step you should be able to answer:
- What are the required claims in an Oath Token?
- What does the verification sequence look like?
- What is the `om+jwt` type header for?

### Step 3: Understand the Architecture (10 minutes)

Read:
1. `.antigravity/memory/project.md` — architecture overview, module map, core abstractions
2. `.antigravity/memory/entities.md` — components, systems, and competitors
3. `.antigravity/decisions/ADR-002-auth-strategy.md` — auth model
4. `.antigravity/decisions/ADR-003-tech-stack.md` — language choices

After this step you should be able to answer:
- What language is each module written in?
- Where does each component live in the monorepo?
- How does authentication differ from authorization in OathMesh?

### Step 4: Understand the Rules (5 minutes)

Read:
1. `.antigravity/rules/core.md` — non-negotiable rules, priority hierarchy, performance budgets
2. `.antigravity/rules/coding-standards.md` — skim the section for the language you'll be working in
3. `.antigravity/rules/security-redlines.md` — actions requiring human approval

### Step 5: Understand Current State (5 minutes)

Read:
1. `.antigravity/context/current-phase.md` — active phase, priorities, blockers, off-limits zones

### Step 6: First Task

Read `.antigravity/onboarding/first-task.md` for a safe, scoped, high-signal first task.

## Graduation Criteria

Onboarding is complete when you can:

- [ ] Explain OathMesh's core value proposition in one sentence
- [ ] Name all six required token claims
- [ ] Describe the authentication/authorization separation
- [ ] Identify which module a given task belongs to
- [ ] Navigate the `.antigravity/` directory tree without guidance
- [ ] Know where to look before making an architectural change
- [ ] Know what requires human approval before proceeding

Total estimated time: 35 minutes.
