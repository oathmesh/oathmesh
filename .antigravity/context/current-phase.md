---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-12"
owner: "Lead Agent"
purpose: "Active sprint, priorities, blockers, and off-limits zones — update frequently"
review_by: "2026-04-19"
---

# OathMesh — Current Phase

## Active Phase

**Phase 1 — MVP Build**

Goal: First working product with GitHub Actions → Issuer → Internal API demo.

## Completed Phase 0 Exit Criteria

- [x] Architect Agent signed off on package structure and dependency graph
- [x] Security Agent approved crypto/ed25519 (stdlib) as signing primitive
- [x] internal/core has zero external imports (verified by `go list -deps`)
- [x] All 3 ADRs written and committed to docs/decisions/
- [x] Protocol spec is frozen - any change requires new ADR
- [x] OathMeshError covers all hard rejection conditions from Part 3
- [x] AuditEvent struct matches audit schema exactly
- [x] go build ./... passes with zero errors
- [x] go vet ./... passes with zero warnings

## Phase 0 Deliverables Completed (2026-04-12)

| # | Deliverable | Status |
|---|-------------|--------|
| 1 | Finalize claim schema | ✅ Done in token.go |
| 2 | internal/core/errors.go | ✅ Done |
| 3 | internal/core/audit.go | ✅ Done |
| 4 | internal/core/context.go | ✅ Done |
| 5 | internal/core/replay.go | ✅ Done |
| 6 | policy/schema.pkl | ✅ Done |
| 7 | policy/example.pkl | ✅ Done |
| 8 | internal/config/issuer.pkl | ✅ Done |
| 9 | go.mod (module github.com/oathmesh/oathmesh; go 1.22) | ✅ Done |
| 10 | .env.example | ✅ Done |
| 11 | .gitignore | ✅ Done |
| 12 | docker-compose.yml | ✅ Done |
| 13 | Root README.md | ✅ Done |
| 14 | ADR-001: Module and package structure | ✅ Done |
| 15 | ADR-002: Technology stack | ✅ Done |
| 16 | ADR-003: Crypto/ed25519 vs alternatives | ✅ Done |

## Known Blockers

| Blocker | Owner | Status | Impact |
|---|---|---|---|
| None | - | - | - |

## Next Phase

**Phase 1 — MVP Build** (starts after Phase 0 exit criteria met)

Goal: First working product with GitHub Actions → Issuer → Internal API demo.

Deliverables: Issuer service, JWKS, CLI, Node middleware, Python middleware, GitHub Actions example, example API, example policy, local audit logging.

Exit criteria: Working end-to-end demo in under 15 minutes.

## Milestone Timeline

| Phase | Target Start | Target End | Status |
|---|---|---|---|
| Phase 0 — Protocol Freeze | 2026-04-05 | 2026-04-12 | **Completed** |
| Phase 1 — MVP Build | 2026-04-12 | 2026-06-14 | **Active** |
| Phase 2 — Developer Product | 2026-06-14 | 2026-08-09 | Planned |
| Phase 3 — Team Product | 2026-08-09 | 2026-10-04 | Planned |
