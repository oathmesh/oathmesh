---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Bug investigation and fix workflow"
---

# Workflow: Bug Fix

## Entry Conditions

- A bug has been reported or discovered
- Unintended behavior differs from specification or expected output
- A test is failing unexpectedly

## Steps

### 1. Reproduce

- Activate **Debug Mode** persona
- Identify the exact conditions that trigger the bug
- Document: expected behavior vs. actual behavior
- Create a minimal reproduction case if possible
- If not reproducible, document what was tried and move to step 6

### 2. Isolate

- Narrow down to the specific module, file, and function
- Use `git log --oneline -10 -- <file>` to check recent changes
- Use `git bisect` if the bug was recently introduced
- Check if the bug is in OathMesh code or a dependency

### 3. Hypothesize

- Form 2–3 specific hypotheses about the root cause
- Reference `personas/debug-mode.md` § OathMesh-Specific Debug Patterns
- Check each hypothesis against the relevant skill file

### 4. Fix

- Switch to **Senior Engineer** persona for implementation
- Write the minimal fix that addresses the root cause
- Follow `rules/coding-standards.md` for the relevant language
- Check `security/redlines.md` if the fix touches auth/crypto/policy code

### 5. Verify

- Activate **Test Agent**
- Write a regression test that would have caught this bug
- Verify the fix resolves the original reproduction case
- Verify no other tests are broken by the fix
- Check for similar patterns elsewhere: "Does this bug exist in other places?"

### 6. Document

- Write a clear commit message explaining the bug and the fix
- If the bug was caused by an ambiguity in `oathmesh.txt`, flag it for spec clarification
- Update `memory/project.md` § Known Technical Debt if the fix is partial or introduces debt
- Log in session memory

## Exit Criteria

- [ ] Bug is reproduced (or documented as non-reproducible)
- [ ] Root cause is identified
- [ ] Fix is implemented
- [ ] Regression test exists and passes
- [ ] All existing tests pass
- [ ] Commit message explains the root cause and fix
