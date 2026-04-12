---
version: "2.0"
created: "2026-04-05"
last_modified: "2026-04-12"
owner: "Founder"
purpose: "OathMesh AI environment — read this first at every session start"
---

# OathMesh AI Environment

This folder is the persistent AI brain for the OathMesh project. It is read by Kilo.ai at the start of every session and updated as the project evolves.

**If you are an AI assistant: load the files in the Boot Sequence below before taking any action. Do not skip steps. Do not write code before completing the boot sequence.**

---

## One-Line Positioning

> **OathMesh gives every machine call a short-lived signed identity.**

## Core Doctrine (say this everywhere)

> **OathMesh authenticates the caller. The receiver authorizes the request.**

---

## Boot Sequence

Load context in this exact order. Stop and report if any file is missing or unreadable.

```
1. This file (README.md)              — orientation and system map
2. context/system-prompt.md           — full Kilo.ai behavioral contract (the authoritative spec)
3. context/current-phase.md           — active phase, completed deliverables, blockers, next action
4. rules/core.md                      — non-negotiable rules and NEVER list
5. rules/security-redlines.md         — actions that always require human approval
6. memory/project.md                  — ADR decisions, architecture state, tech debt
7. memory/entities.md                 — key components, packages, owners
8. docs/decisions/ (latest ADRs)      — check before suggesting any architectural change
```

**After boot:** state which phase you are in and which exit criteria remain before proposing any work.

---

## What Is OathMesh

OathMesh is a micro-protocol and developer product that replaces shared machine secrets (API keys, static tokens) with short-lived, signed JWT-based call identity.

**Built for:** agents, CI/CD jobs, internal tools, automation bots, MCP-adjacent tool backends, service-to-service calls.

**Not:** a user auth system, a service mesh, a general IAM platform, a replacement for SPIFFE.

### Core Components

| Component | Description |
|---|---|
| Issuer | Creates and signs Oath Tokens (short-lived JWTs) |
| Caller | Agent, workflow, bot, or service making the request |
| Receiver | API or service accepting the request |
| Verifier | Middleware that validates the Oath Token (14 steps) |
| Policy Engine | Evaluates whether the verified caller is allowed — Pkl-backed |
| Audit Pipeline | Records every allow and deny — NDJSON, non-optional |

### Golden Path (MVP wedge)
```
GitHub Actions → OathMesh Issuer → Internal API
```

---

## Technology Stack

| Concern | Choice | Why |
|---|---|---|
| Language | Go 1.22+ | stdlib crypto, fast binary, infra ecosystem |
| Config DSL | Apple Pkl | typed policies, IDE validation, `pkl-gen-go` |
| HTTP framework | chi/v5 | lightweight, idiomatic, stdlib-compatible |
| Signing | `crypto/ed25519` (stdlib) | no CVE surface from third-party JWT libs |
| Audit | NDJSON via `log/slog` | structured, zero-dep, cloud-native |
| Policy hot-reload | `fsnotify` | cross-platform, zero-downtime |
| Replay cache (dev) | `MemoryReplayCache` + `sync.RWMutex` | zero deps for local use |
| Replay cache (prod) | Redis via `go-redis/v9` | atomic SET NX, horizontally scalable |

**Module path (canonical):** `github.com/oathmesh/oathmesh`

**Pkl note:** Use `amends "schema.pkl"` (local relative path). The remote `package://pkg.pkl-lang.org/oathmesh/...` URI does not exist yet — never use it.

---

## Directory Map

```
.antigravity/
├── README.md                          ← You are here. Read first.
│
├── context/
│   ├── system-prompt.md               ← Full Kilo.ai system prompt — the authoritative spec.
│   │                                    Contains: identity, stack, protocol, agents, phases,
│   │                                    NEVER list, behavioral rules, session protocol.
│   └── current-phase.md               ← Active phase, deliverables done, blockers, next action.
│
├── rules/
│   ├── core.md                        ← Non-negotiable rules. Design principles in priority order.
│   ├── coding-standards.md            ← Go + Pkl naming, formatting, package conventions.
│   ├── security-redlines.md           ← Actions requiring human approval before proceeding.
│   └── conflict-resolution.md         ← How to resolve contradictions between rules or agents.
│
├── memory/
│   ├── project.md                     ← ADR decisions, architecture choices, tech debt log.
│   ├── entities.md                    ← Key components, packages, owners, current state.
│   └── expiry-policy.md               ← When memory entries go stale and what to do.
│
└── CHANGELOG.md                       ← All changes to this configuration folder.
```

---

## Rule Hierarchy

When rules conflict, resolve in this fixed priority order:

1. **Security** — `rules/security-redlines.md` and Security Agent hard redlines in `context/system-prompt.md`
2. **Protocol** — `docs/decisions/` ADRs and protocol reference in `context/system-prompt.md` Part 3
3. **Architecture** — `rules/core.md` and Architect Agent
4. **Style** — `rules/coding-standards.md`

Full resolution: `rules/conflict-resolution.md`

---

## Agent Map

| Agent | Owns | Cannot cross into |
|---|---|---|
| Lead (Kilo.ai session) | Session state, phase gating, ADR coordination | `internal/sign/`, `internal/verify/` |
| Architect | Package structure, dependency graph, `ARCHITECTURE.md` | Signing/crypto code |
| Security | `internal/sign/`, `internal/verify/`, key management | Feature code |
| Test | Test contracts, coverage, `go test -race` | Protocol spec |
| Docs | All public-facing content, CLI reference, examples | Architecture decisions |

**Crossing a boundary requires explicit sign-off from the owning agent.**

---

## Phase Summary

| Phase | Week | Focus |
|---|---|---|
| 0 | 1 | Protocol freeze, repo scaffold, ADRs, Pkl schemas |
| 1 | 2 | Issuer service, signing, GitHub OIDC exchange |
| 2 | 3 | Verifier (14 steps), Go/Python/Node SDKs |
| 3 | 4 | Pkl policy engine, audit pipeline, replay cache |
| 4 | 5 | CLI (`mint`, `verify`, `inspect`, `serve`, `keys rotate`, `policy validate`) |
| 5 | 5 | Gateway mode (parallel with CLI) |
| 6 | 6 | End-to-end demo, GitHub Actions, docker-compose |
| 7 | 7 | Full documentation (zero placeholder text) |
| 8 | 8 | Polish, DX, launch prep, CI pipeline |

**Rule:** do not start any phase until ALL exit criteria of the previous phase are confirmed with ✓.

---

## Security Contract (summary — full detail in `rules/security-redlines.md`)

The following are hard stops. No exception. No "this time it's fine":

```
✗ Private key — never in logs, never in responses, never hardcoded
✗ Full Oath Token string — never in logs
✗ TTL > 300 seconds — server enforces, callers cannot override
✗ exp set by caller — always ignored; server sets it
✗ Algorithm "none" — rejected at Step 02 before any other check
✗ Unknown issuer — explicit allowlist only, no wildcards
✗ Default allow — always default deny; never configurable
✗ YAML policy files — Pkl only
✗ Third-party JWT libraries — we own the signing path
✗ http.DefaultClient for JWKS — always 5-second timeout client
```

If you are uncertain whether an action crosses a redline: stop, surface the question, wait for human confirmation.

---

## Pkl Workflow

```
1. Edit policy file        → policy/production.pkl (amends "schema.pkl")
2. Validate locally        → pkl eval policy/production.pkl
3. Schema change?          → edit policy/schema.pkl (requires ADR)
4. After schema change     → make pkl-gen (regenerates internal/policy/generated.go)
5. Commit generated.go     → always commit alongside schema change
6. Deploy                  → copy .pkl file to server; fsnotify hot-reloads without restart
```

---

## Session Start Checklist

Before writing any code, confirm:
- [ ] Boot sequence loaded (all 8 files)
- [ ] Current phase identified from `context/current-phase.md`
- [ ] Remaining exit criteria for current phase reviewed
- [ ] No open ADR conflicts detected in `docs/decisions/`
- [ ] If working on `internal/sign/` or `internal/verify/`: Security Agent activated

---

## Session End Checklist

Before closing the session:
- [ ] Session log written to `docs/session-logs/YYYY-MM-DD.md`
- [ ] `context/current-phase.md` updated with progress
- [ ] New ADRs committed to `docs/decisions/` if architectural decisions were made
- [ ] `ARCHITECTURE.md` updated if package structure changed
- [ ] `memory/project.md` updated if key decisions were made
- [ ] Handoff summary output in exact format from `context/system-prompt.md` Part 10

---

## Emergency Procedures

| Situation | Action |
|---|---|
| Rules conflict | `rules/conflict-resolution.md` |
| Uncertain about a security action | `rules/security-redlines.md` then ask the human |
| ADR conflict with code | Surface it — never silently resolve |
| Pkl schema error at load time | Fix the schema, re-run `make pkl-gen`, commit |
| Missing context file | Stop, report which file is missing, do not proceed |
| Context window exhausted | Re-read Part 3 (protocol) and current phase before resuming |

---

## Repo Quick Reference

```
github.com/oathmesh/oathmesh

internal/core/      ← zero external deps — protocol types, errors, audit, replay interface
internal/sign/      ← Security Agent only — Ed25519, JWS
internal/verify/    ← Security Agent only — 14 verification steps, JWKS cache, replay
internal/policy/    ← Pkl engine, fsnotify hot-reload
internal/audit/     ← AuditSink interface + implementations
internal/issuer/    ← HTTP issuer (chi), mint, exchange, JWKS, discovery
internal/gateway/   ← Reverse proxy mode
internal/config/    ← Config struct, Pkl + env merge
sdk/go/             ← Go middleware + CallerFrom helper
sdk/python/         ← Python verify + FastAPI integration
sdk/node/           ← TypeScript middleware + types
policy/             ← schema.pkl, example.pkl, production.pkl
examples/           ← chi-api, express-api, fastapi-api, github-actions, curl
docs/               ← decisions/, protocol/, quickstarts/, config/, security/, session-logs/
```
