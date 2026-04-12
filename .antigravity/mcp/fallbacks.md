---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Fallback chain when MCP servers are unavailable"
---

# MCP Server Fallbacks

When an MCP server is unavailable, use these fallback strategies. Never silently proceed without the capability — acknowledge the degraded state and adapt.

## Fallback Chain by Server

### Filesystem Server (REQUIRED)

**If unavailable:** Use the terminal server to run file operations:
- Read: `cat <path>`, `type <path>` (Windows)
- Write: redirect output, or use the AI's built-in file writing capability
- List: `ls -la`, `dir` (Windows), `find . -type f`
- Search: `grep -r`, `findstr` (Windows)

**If terminal is also unavailable:** Operate in read-only advisory mode. Report to human that no file operations are possible.

### Git Server (REQUIRED)

**If unavailable:** Use the terminal server to run git commands directly:
- `git log --oneline -20`
- `git diff HEAD`
- `git status`
- `git blame <file>`

**If terminal is also unavailable:** Operate without version control context. Flag that commit history and diff checks are skipped.

### Terminal Server (REQUIRED)

**If unavailable:** This is a foundational server. Without it:
- Cannot run builds, tests, or linting
- Cannot install dependencies
- Cannot execute project commands

**Action:** Halt active work. Switch to advisory-only mode:
- Review and comment on code
- Draft documentation
- Create ADRs and plans
- Do NOT generate code that cannot be tested

Report: "Terminal server unavailable. Operating in advisory mode. Cannot build, test, or execute commands."

### Browser Server (OPTIONAL)

**If unavailable:** Use the search server for information retrieval. If specific documentation pages are needed, ask the human to:
1. Navigate to the URL
2. Copy the relevant content
3. Paste it into the conversation

**Degraded capability:** Cannot take screenshots, cannot interact with web UIs, cannot test HTTP endpoints visually.

### Search Server (OPTIONAL)

**If unavailable:** Use the browser server to navigate directly to known URLs:
- Go documentation: `https://pkg.go.dev`
- Node.js docs: `https://nodejs.org/docs`
- RFC documents: `https://datatracker.ietf.org`
- JWT debugger: `https://jwt.io`
- GitHub Actions docs: `https://docs.github.com/en/actions`

If both search and browser are unavailable, rely on built-in knowledge and flag any uncertainty: "Unable to verify against current documentation — this answer uses training data which may be outdated."

### Memory Server (OPTIONAL)

**If unavailable:** Use the filesystem server to read/write memory files directly in `.antigravity/memory/`:
- Read: load `memory/project.md`, `memory/entities.md`
- Write: update files directly using filesystem operations
- This is the designed fallback — the memory files ARE the persistent memory

**Degraded capability:** No cross-session key-value store. All memory is file-based only.

## Degraded Mode Classification

| Missing Servers | Mode | Capabilities |
|---|---|---|
| None | **Full** | All capabilities available |
| Browser + Search | **Standard** | Code, test, deploy — no web research |
| Memory | **Standard** | All capabilities — memory via filesystem |
| Browser + Search + Memory | **Core** | Code, test, deploy — no research, filesystem memory |
| Terminal | **Advisory** | Review, document, plan — no execution |
| Filesystem | **Advisory** | Read-only conversation — no project interaction |
| Filesystem + Terminal | **Offline** | Conversation only — no project access |

## Reporting Degraded State

When entering a degraded mode, output this notice at the start of any work:

```
⚠ DEGRADED MODE: [mode name]
   Unavailable: [list of unavailable servers]
   Impact: [what cannot be done]
   Fallback: [what is being used instead]
```

Continue checking server availability periodically. When a server comes back, log the recovery and resume full capabilities.
