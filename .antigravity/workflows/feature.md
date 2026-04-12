---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Feature development workflow — discovery through documentation"
---

# Workflow: Feature Development

## Entry Conditions

- A feature request or specification section is being implemented
- The feature does not exist yet (not a bug fix or refactor)
- The active phase in `context/current-phase.md` includes this feature

## Steps

### 1. Discover

- Read the relevant section of `oathmesh.txt` (the source-of-truth specification)
- Check `decisions/` for ADRs that constrain this feature
- Check `memory/project.md` for related modules and abstractions
- Check `context/current-phase.md` for blockers relevant to this feature
- Identify affected modules and components

**Artifacts:** Understanding of what to build and what constraints apply.

### 2. Design

- Activate **Architect Agent** if the feature creates new components or changes interfaces
- Define the public API surface (function signatures, endpoints, CLI commands)
- Identify data structures needed (reference `skills/data-modeling.md`)
- Check performance budgets in `rules/core.md`
- If a significant design decision is needed, write an ADR using `decisions/ADR-000-template.md`

**Artifacts:** API design, data structures, and optionally a new ADR.

### 3. Implement

- Activate **Senior Engineer** persona
- Follow `rules/coding-standards.md` for the relevant language
- Write code in the correct module per `rules/core.md` § AC-5
- Handle errors with structured error types per `personas/voice.md` § Error Message Voice
- Check `security/redlines.md` before any security-sensitive operation
- Commit frequently with conventional commit messages (`tools/git.md`)

**Artifacts:** Working implementation code.

### 4. Test

- Activate **Test Agent**
- Write tests alongside implementation (not after)
- Meet quality gates from `agents/test-agent.md`
- Write unit tests for all public functions
- Write integration tests for cross-component interactions
- Verify against performance budgets

**Artifacts:** Test files, coverage report.

### 5. Document

- Activate **Docs Agent**
- Write/update public API documentation
- Update code examples if API changed
- Update `memory/project.md` if architecture changed
- Update `memory/entities.md` if new components added
- Add terms to `context/glossary.md` if new concepts introduced

**Artifacts:** Updated documentation, memory files.

## Exit Criteria

- [ ] Implementation complete and compiles/passes type checking
- [ ] All tests pass with required coverage
- [ ] Linting clean (zero warnings)
- [ ] Documentation updated
- [ ] Memory updated (if applicable)
- [ ] No security redlines violated
- [ ] Ready for review (`workflows/review.md`)

## Rollback Conditions

If the feature cannot be completed:
- Document what was attempted and why it failed
- Revert any incomplete changes
- Update `context/current-phase.md` with the blocker
- Log in session memory
