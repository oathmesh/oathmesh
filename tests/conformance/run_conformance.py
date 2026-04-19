#!/usr/bin/env python3
"""Lightweight cross-SDK conformance harness."""

from __future__ import annotations

import json
import os
import subprocess
import sys
import time
import shutil
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
CASES_PATH = ROOT / "tests" / "conformance" / "cases.json"
OUTPUT_DIR = ROOT / "tests" / "conformance" / "results"
OUTPUT_PATH = OUTPUT_DIR / "conformance_matrix.json"


def _tail(text: str, lines: int = 20) -> str:
    parts = text.strip().splitlines()
    return "\n".join(parts[-lines:])


def main() -> int:
    if not CASES_PATH.exists():
        print(f"[error] missing cases file: {CASES_PATH}")
        return 1

    spec = json.loads(CASES_PATH.read_text(encoding="utf-8"))
    targets = spec["targets"]
    cases = spec["cases"]

    matrix: list[dict] = []
    failures = 0

    print("=== OathMesh Conformance Harness ===")
    print(f"cases: {len(cases)} | targets: {', '.join(targets)}")

    for case in cases:
        case_id = case["id"]
        row = {"id": case_id, "category": case["category"], "description": case["description"], "targets": {}}
        print(f"\n[case] {case_id}")

        for target in targets:
            check = case["checks"].get(target, {})
            if "skip" in check:
                reason = check["skip"]
                row["targets"][target] = {"status": "skipped", "reason": reason}
                print(f"  - {target}: SKIP ({reason})")
                continue

            cmd = check["command"]
            if cmd and os.name == "nt" and "." not in cmd[0]:
                cmd_cmd = f"{cmd[0]}.cmd"
                if shutil.which(cmd_cmd):
                    cmd = [cmd_cmd, *cmd[1:]]
            cwd = ROOT / check.get("cwd", ".")
            started = time.time()
            proc = subprocess.run(
                cmd,
                cwd=str(cwd),
                capture_output=True,
                text=True,
                shell=False,
                check=False,
            )
            elapsed_ms = int((time.time() - started) * 1000)
            passed = proc.returncode == 0
            status = "pass" if passed else "fail"
            if not passed:
                failures += 1
            row["targets"][target] = {
                "status": status,
                "returncode": proc.returncode,
                "command": cmd,
                "cwd": str(cwd),
                "elapsed_ms": elapsed_ms,
                "stdout_tail": _tail(proc.stdout),
                "stderr_tail": _tail(proc.stderr),
            }
            print(f"  - {target}: {'PASS' if passed else 'FAIL'} ({elapsed_ms} ms)")
        matrix.append(row)

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    OUTPUT_PATH.write_text(
        json.dumps(
            {
                "generated_at_unix": int(time.time()),
                "spec_version": spec.get("version"),
                "targets": targets,
                "cases": matrix,
                "failed_checks": failures,
            },
            indent=2,
        )
        + "\n",
        encoding="utf-8",
    )

    print("\n=== Conformance Matrix ===")
    header = f"{'case_id':45} {'go':8} {'node':8} {'python':8}"
    print(header)
    print("-" * len(header))

    def icon(status: str) -> str:
        if status == "pass":
            return "PASS"
        if status == "fail":
            return "FAIL"
        return "SKIP"

    for row in matrix:
        go_s = icon(row["targets"]["go"]["status"])
        node_s = icon(row["targets"]["node"]["status"])
        py_s = icon(row["targets"]["python"]["status"])
        print(f"{row['id'][:45]:45} {go_s:8} {node_s:8} {py_s:8}")

    print(f"\nresults: {OUTPUT_PATH}")
    if failures:
        print(f"conformance: FAIL ({failures} failed checks)")
        return 1

    print("conformance: PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
