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
- `alg=none` rejection
- subject format validation
- binding-required semantics (`requireRequestBinding` / `require_request_binding`)
- future `iat` rejection
- revocation expectations (Go core; marked SKIP/N/A for Node/Python)
- middleware auth header handling semantics

Each case points to existing test tooling commands (`go test`, `vitest`, `pytest`).

## Current case matrix

| Case ID | Go | Node | Python |
|---|---|---|---|
| `token_parsing_validation_failures` | ✅ | ✅ | ✅ |
| `issuer_check_untrusted` | ✅ | ✅ | ✅ |
| `audience_check_mismatch` | ✅ | ✅ | ✅ |
| `replay_detection_semantics` | ✅ | ✅ | ✅ |
| `revocation_subject_revoked` | ✅ | SKIP (no revocation list wiring in conformance target) | SKIP (no revocation list wiring in conformance target) |
| `alg_none_rejection` | ✅ | ✅ | ✅ |
| `subject_format_validation` | ✅ | ✅ | ✅ |
| `binding_required_semantics` | ✅ | ✅ | ✅ |
| `iat_future_rejection` | ✅ | ✅ | ✅ |
| `middleware_auth_header_handling_semantics` | ✅ | ✅ | ✅ |

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
