#!/usr/bin/env python3
"""Validate bench/competitors.json against the required schema shape."""
import json
import os
import sys

REQUIRED_TOP = {"measured_on", "tool", "verdict_rows"}
REQUIRED_TOOL = {
    "runtime_us_mean", "runtime_us_min", "runtime_us_max", "runs",
    "needs_root", "install_size_bytes", "output_bytes", "notes", "source",
}
REQUIRED_MEASURED = {"host", "kernel", "cpus", "date"}
REQUIRED_VERDICT = {"capability", "asbuilt", "competitor", "source"}
MUST_INCLUDE = {"asbuilt", "lshw", "dmidecode", "lscpu", "lspci",
                "perfspect", "pcicrawler", "sat-hwmatch"}

path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "competitors.json")
with open(path, encoding="utf-8") as fh:
    doc = json.load(fh)

errors = []

missing_top = REQUIRED_TOP - set(doc)
if missing_top:
    errors.append(f"missing top-level keys: {sorted(missing_top)}")

m = doc.get("measured_on", {})
missing_m = REQUIRED_MEASURED - set(m)
if missing_m:
    errors.append(f"measured_on missing: {sorted(missing_m)}")

tools = doc.get("tool", {})
missing_required = MUST_INCLUDE - set(tools)
if missing_required:
    errors.append(f"tool missing required entries: {sorted(missing_required)}")

for name, entry in tools.items():
    missing = REQUIRED_TOOL - set(entry)
    if missing:
        errors.append(f"tool[{name}] missing: {sorted(missing)}")
        continue
    if not isinstance(entry["needs_root"], bool):
        errors.append(f"tool[{name}].needs_root must be bool")
    for k in ("runtime_us_mean", "runtime_us_min", "runtime_us_max",
              "runs", "install_size_bytes", "output_bytes"):
        v = entry[k]
        if v is not None and not isinstance(v, int):
            errors.append(f"tool[{name}].{k} must be int or null, got {type(v).__name__}")
    # mean must sit between min and max
    mean, lo, hi = (entry["runtime_us_mean"], entry["runtime_us_min"], entry["runtime_us_max"])
    if None not in (mean, lo, hi) and not (lo <= mean <= hi):
        errors.append(f"tool[{name}]: mean {mean} not within [{lo}, {hi}]")
    if mean is not None and entry["runs"] < 1:
        errors.append(f"tool[{name}]: has timings but runs={entry['runs']}")
    if not entry["notes"].strip():
        errors.append(f"tool[{name}]: empty notes")

for i, row in enumerate(doc.get("verdict_rows", [])):
    missing = REQUIRED_VERDICT - set(row)
    if missing:
        errors.append(f"verdict_rows[{i}] missing: {sorted(missing)}")

if errors:
    print("FAIL")
    for e in errors:
        print("  -", e)
    sys.exit(1)

print("PASS: competitors.json is valid")
print(f"  tools: {len(tools)} -> {', '.join(sorted(tools))}")
print(f"  verdict_rows: {len(doc['verdict_rows'])}")
measured = [n for n, e in tools.items() if e["runtime_us_mean"] is not None]
sourced = [n for n, e in tools.items() if e["runtime_us_mean"] is None]
print(f"  measured ({len(measured)}): {', '.join(sorted(measured))}")
print(f"  not measurable ({len(sourced)}): {', '.join(sorted(sourced))}")
