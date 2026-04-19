# Conformance Framework (Phase 2: `p2-conformance-framework`)

This suite validates behavior parity for key verification and middleware semantics across:

- Go core verifier (`internal/verify`)
- Node SDK (`sdk/node`)
- Python SDK (`sdk/python`)

## What it checks

Canonical cases are defined in `tests/conformance/cases.json`:

- token parsing/validation failures
- issuer trust checks
- audience checks
- replay detection
- revocation expectations (Go core; marked N/A for Node/Python)
- middleware auth header handling semantics

Each case points to existing test tooling commands (`go test`, `vitest`, `pytest`).

## Run locally

From repository root:

```bash
python tests/conformance/run_conformance.py
```

or:

```bash
make conformance
```

## Outputs

Results are written to:

- `tests/conformance/results/conformance_matrix.json`

The harness prints a pass/fail/skip matrix and exits non-zero if any non-skipped check fails.

## CI

The CI conformance job runs `make conformance` after installing Go/Node/Python SDK test dependencies.
