---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Code review workflow — checklist, ADR alignment, security scan, style gate"
---

# Workflow: Code Review

## Entry Conditions

- Code is ready for review (implementation and tests complete)
- `workflows/feature.md` or `workflows/bugfix.md` exit criteria met

## Steps

### 1. Activate Review Mode

- Switch to **Review Mode** persona
- Load the review checklist from `personas/review-mode.md`

### 2. Correctness Review

- [ ] Does the code do what the specification or task requires?
- [ ] Are all edge cases handled?
- [ ] Are error paths complete and structured?
- [ ] Do error messages follow `personas/voice.md` § Error Message Voice?

### 3. Architecture Alignment

- [ ] Check against ALL active ADRs in `decisions/`
- [ ] Does this maintain authentication vs. authorization separation (ADR-002)?
- [ ] Does this keep the issuer stateless (AC-4 in `rules/core.md`)?
- [ ] Is code in the correct module (AC-5 in `rules/core.md`)?
- [ ] If this is a new component, was an ADR written?

### 4. Security Scan

- [ ] Activate **Security Agent** checks from `agents/security-agent.md`
- [ ] No secrets in code
- [ ] No full tokens logged
- [ ] Crypto operations use approved algorithms and libraries
- [ ] Token verification follows the 14-step sequence from `skills/auth.md`
- [ ] Policy evaluation is default-deny
- [ ] If gateway code: incoming `X-Oathmesh-*` headers stripped before verification

### 5. Test Assessment

- [ ] Activate **Test Agent** quality gates from `agents/test-agent.md`
- [ ] Coverage threshold met (≥ 80% new code, ≥ 90% for verification/policy)
- [ ] All tests pass
- [ ] Security-critical paths at 100% coverage
- [ ] Error codes have corresponding tests

### 6. Style Gate

- [ ] Follows `rules/coding-standards.md` for the relevant language
- [ ] Terminology matches `context/glossary.md`
- [ ] Linting clean (zero warnings)
- [ ] Type checking passes (strict mode)

### 7. Documentation Check

- [ ] Public API documented?
- [ ] Examples updated if API changed?
- [ ] `memory/project.md` updated if architecture changed?

### 8. Verdict

Produce one of:

| Verdict | Meaning |
|---|---|
| **Approve** | Ship it. All checks pass. |
| **Approve with changes** | Minor issues that don't block merging. List them. |
| **Request changes** | Issues that must be fixed before merging. List them with severity labels. |

## Exit Criteria

- [ ] All checklist items evaluated
- [ ] Findings documented with severity labels (MUST FIX / SHOULD FIX / CONSIDER / NIT)
- [ ] Verdict issued
