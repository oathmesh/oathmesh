---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Hotfix workflow — fast-path with mandatory post-mortem"
---

# Workflow: Hotfix

## Entry Conditions

- A critical bug in production or a security vulnerability
- The fix cannot wait for a normal `workflows/bugfix.md` cycle
- The human has identified this as urgent

## Steps

### 1. Assess

- What is the blast radius? (Which callers, receivers, or gateways are affected?)
- Is this a security issue? (If yes, Security Agent has veto power on the fix approach)
- What is the temporary mitigation? (Can we disable the feature, block traffic, or roll back?)

### 2. Fix (Minimal)

- Write the smallest possible fix that resolves the immediate issue
- Security redlines STILL APPLY — no shortcuts on crypto, secrets, or auth logic
- Tests can be minimal (happy path + the specific bug) — full test suite comes in step 5
- Use `HOTFIX:` comment prefix to mark expedient choices

### 3. Deploy

- Follow normal deployment gate (human approval per `rules/security-redlines.md`)
- Deploy to staging first if time permits; direct to production only if the outage is ongoing

### 4. Verify

- Confirm the fix resolves the production issue
- Monitor for 30 minutes after deployment

### 5. Post-Mortem (Mandatory)

This step is NOT optional. Within 48 hours of the hotfix:

- [ ] Write a full regression test for the bug
- [ ] Replace any `HOTFIX:` expedient code with production-quality code
- [ ] Write a post-mortem:
  - What happened?
  - Why did it happen?
  - How was it detected?
  - How was it fixed?
  - What prevented it from being caught earlier?
  - What needs to change to prevent recurrence?
- [ ] Update relevant rules, skills, or tests based on post-mortem findings
- [ ] File improvement proposal in `feedback/proposals/` if systemic issue found

## Exit Criteria

- [ ] Production issue resolved
- [ ] Regression test written
- [ ] Post-mortem written within 48 hours
- [ ] `HOTFIX:` code replaced with production-quality code
- [ ] Improvement proposals filed if applicable
