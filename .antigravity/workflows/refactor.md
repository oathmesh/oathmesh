---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Refactoring workflow — scoped, safe, incremental"
---

# Workflow: Refactoring

## Entry Conditions

- Code structure or organization needs improvement without changing external behavior
- Technical debt is being addressed (from `memory/project.md` § Known Technical Debt)
- The refactoring is planned and scoped (not ad-hoc cleanup during feature work)

## Steps

### 1. Scope Definition

- Define exactly what will change and what will NOT change
- Identify the affected files and modules
- Confirm external behavior will not change (API surface, error codes, config options)
- Check `context/current-phase.md` for off-limits zones
- Check `decisions/` for ADRs that constrain the refactoring

### 2. Safety Net

Before making any changes:

- [ ] All existing tests pass (green baseline)
- [ ] Test coverage is documented for the affected area
- [ ] Git working tree is clean (commit or stash everything)
- [ ] Create a checkpoint commit or branch: `refactor/<brief-description>`

### 3. Incremental Execution

Make changes in small, independent steps. Each step must:

- Compile and pass all tests before proceeding to the next
- Be committable on its own with a meaningful commit message
- Not mix structural changes with behavioral changes

Order of operations:
1. Rename (files, functions, types) — one rename per commit
2. Move (relocate code to correct module) — one move per commit
3. Extract (pull out new abstractions) — one extraction per commit
4. Simplify (reduce complexity) — one simplification per commit

### 4. Verification

After all refactoring steps:

- [ ] All existing tests pass (same green as baseline)
- [ ] No new warnings from linting
- [ ] Type checking passes
- [ ] External API surface is unchanged (same function signatures, endpoints, CLI commands)
- [ ] Performance budgets still met (run benchmarks if applicable)

### 5. Documentation Update

- [ ] Update code comments if module structure changed
- [ ] Update `memory/project.md` if module map changed
- [ ] No documentation changes needed for external-facing docs (behavior unchanged)

### 6. Review

- Submit for review using `workflows/review.md`
- Reviewer should confirm: "The external behavior is unchanged"

## Exit Criteria

- [ ] All tests pass (same as pre-refactoring baseline)
- [ ] External API surface unchanged
- [ ] Each step is a clean, meaningful commit
- [ ] Memory and documentation updated if internal structure changed
- [ ] Review approved

## Anti-Patterns to Avoid

- Mixing refactoring with feature work in the same session
- Refactoring without a green test baseline
- Making all changes in a single large commit
- Changing public API during a refactoring (that's a feature, use `workflows/feature.md`)
- Refactoring code in off-limits zones
