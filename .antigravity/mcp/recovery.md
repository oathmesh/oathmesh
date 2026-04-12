---
version: "1.0"
created: "2026-04-05"
last_modified: "2026-04-05"
owner: "Founder"
purpose: "Recovery playbooks for MCP server failures and system corruption"
---

# MCP Recovery Playbook

## Failure Class 1: Single MCP Server Down

### Symptoms
- One server fails to respond within its configured timeout
- Health check for one server fails
- Operations routed to one server return errors

### Recovery Steps

1. **Identify** which server is down (check `mcp/servers.json` for the server list)
2. **Check fallback** in `mcp/fallbacks.md` — switch to the designated fallback
3. **Report** the degraded state to the human (use the degraded mode format from `mcp/fallbacks.md`)
4. **Retry** after 60 seconds — many MCP server failures are transient
5. **If persistent** (3+ retries over 5 minutes), ask the human to restart the server

### Server-Specific Recovery

| Server | Common Cause | Fix |
|---|---|---|
| Filesystem | Permission error, path not found | Verify workspace path in `servers.json` matches actual project location |
| Git | Not a git repository, corrupted index | Run `git status` via terminal; if needed, `git fsck` |
| Terminal | Shell process crashed | Restart the terminal server process |
| Browser | Puppeteer/Chrome crash | Kill orphaned Chrome processes, restart browser server |
| Search | API key expired or rate limited | Check `OATHMESH_BRAVE_SEARCH_KEY` env var, wait for rate limit reset |
| Memory | Memory file corruption | See Failure Class 3 below |

## Failure Class 2: Multiple MCP Servers Down

### Symptoms
- Two or more servers fail simultaneously
- System cannot perform basic operations

### Recovery Steps

1. **Classify degraded mode** per `mcp/fallbacks.md` table
2. **If Core or Advisory mode** — continue with limited capabilities, focus on planning and review
3. **If Offline mode** — halt all work, inform the human immediately
4. **Check root cause** — common causes of multi-server failure:
   - Node.js/npx not installed or not in PATH
   - Network connectivity lost (affects search, browser)
   - System resource exhaustion (memory, CPU)
5. **Restart in initialization order** per `mcp/servers.json`:
   1. Filesystem
   2. Git
   3. Terminal
   4. Search
   5. Browser
   6. Memory

## Failure Class 3: Memory Corruption

### Symptoms
- `memory/project.md` or `memory/entities.md` contains contradictory information
- Memory entries reference components or decisions that don't exist
- AI behavior becomes inconsistent with known project state

### Recovery Steps

1. **Stop** — do not write additional memory until corruption is resolved
2. **Identify** the corrupted file(s)
3. **Check git history** for the last known-good version:
   ```
   git log --oneline -10 -- .antigravity/memory/
   ```
4. **Compare** the current version with the last known-good version:
   ```
   git diff <last-good-commit> -- .antigravity/memory/project.md
   ```
5. **Restore** from the last known-good commit if the corruption is clear:
   ```
   git checkout <last-good-commit> -- .antigravity/memory/project.md
   ```
6. **If unclear** which version is correct, move the corrupted file to `memory/stale/` with a timestamp suffix and ask the human to validate
7. **Re-read** the source-of-truth files to rebuild memory:
   - `oathmesh.txt` — specification
   - `decisions/ADR-*.md` — architecture decisions
   - Actual source code (if it exists)
8. **Write** a corrected memory file based on the re-read

### Prevention

- Always validate memory writes against source code and ADRs
- Never write memory based solely on conversation — verify against files
- Include `last_modified` and `review_by` dates in all memory files (per `memory/expiry-policy.md`)

## Failure Class 4: Context Overflow

### Symptoms
- AI responses become incoherent or forget recent context
- Session exceeds context window limits
- AI starts contradicting its own earlier statements in the same session

### Recovery Steps

1. **Acknowledge** the overflow — state clearly: "Context limit approaching. Summarizing and resetting."
2. **Write session summary** using `feedback/session-log-template.md`
3. **Save critical state** to `memory/project.md` if architecture changed
4. **Prioritize retention** of (in order):
   1. Current task and its state (what's done, what remains)
   2. Active file paths and their roles
   3. Decisions made in this session
   4. Open questions or blockers
5. **Shed** (in order, first shed = least important):
   1. Full file contents already committed
   2. Research findings already documented
   3. Detailed conversation history
   4. Verbose error logs
6. **Restart** with a fresh context, loading the boot sequence from `README.md`

### Prevention

- Keep individual files under 500 lines (per `rules/coding-standards.md`)
- Summarize findings in memory files rather than keeping raw data in context
- Use `context/current-phase.md` to track state externally rather than in-context

## Failure Class 5: Confidence Threshold Breach

### Symptoms
- AI uncertainty about a decision exceeds the acceptable threshold
- Multiple valid but contradictory approaches exist
- Impact of a wrong choice is high (security, data integrity, protocol compatibility)

### Recovery Steps

1. **Stop** — do not proceed with the uncertain action
2. **State** the uncertainty clearly:
   - What is the decision to be made?
   - What are the options?
   - What is the risk of each option?
   - What information would resolve the uncertainty?
3. **Check** the decision hierarchy:
   - Is this covered by an existing ADR? → Follow the ADR
   - Is this covered by a rule in `rules/core.md`? → Follow the rule
   - Is this a security decision? → Default to the more conservative option and ask human
4. **Escalate** to the human with a clear decision brief
5. **Record** the decision (once made) as an ADR if it constrains future choices

### Confidence Thresholds

| Domain | Threshold | If Below Threshold |
|---|---|---|
| Security and cryptography | 95% confidence required | Always ask human |
| Protocol changes | 90% confidence required | Ask human + require ADR |
| Architecture decisions | 85% confidence required | Ask human |
| Implementation choices | 70% confidence required | Make best choice, flag for review |
| Style and formatting | 50% confidence required | Make best choice, move on |
