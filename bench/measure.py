#!/usr/bin/env python3
"""Benchmark harness for asbuilt vs Linux hardware-audit competitors.

Runs each command N times (default 20) in the same WSL2 environment and
records wall-clock time in microseconds. Output is discarded to DEVNULL so we
measure the tool, not the pipe. Warm-up runs (default 2) are excluded.

Writes raw measurements to bench/raw_measurements.json.
"""
import json
import os
import statistics
import subprocess
import sys
import time

RUNS = int(os.environ.get("BENCH_RUNS", "20"))
WARMUP = int(os.environ.get("BENCH_WARMUP", "2"))

COMMANDS = [
    # label, argv, note
    ("asbuilt-check-fixture", ["/tmp/asbuilt", "--root", "/tmp/server", "check"],
     "full audit of fixture: manifest + 33 rules + text report"),
    ("asbuilt-manifest-fixture", ["/tmp/asbuilt", "--root", "/tmp/server", "manifest"],
     "literal 'manifest' subcommand (text renderer; --json must precede subcommand)"),
    ("asbuilt-manifest-json-fixture", ["/tmp/asbuilt", "--json", "--root", "/tmp/server", "manifest"],
     "true JSON manifest (--json before subcommand)"),
    ("asbuilt-version", ["/tmp/asbuilt", "--version"], "process startup floor"),
    ("asbuilt-check-live", ["/tmp/asbuilt", "check"], "live audit of the WSL machine"),
    ("lshw-json", ["lshw", "-json"], "full hardware tree as JSON"),
    ("lshw-short", ["lshw", "-short"], "short hardware tree"),
    ("dmidecode", ["dmidecode"], "full DMI/SMBIOS dump"),
    ("dmidecode-t17", ["dmidecode", "-t", "17"], "memory device (DIMM) records only"),
    ("lscpu-json", ["lscpu", "-J"], "CPU topology as JSON"),
    ("lspci-vvv", ["lspci", "-vvv"], "PCI device tree with verbose capability data"),
    ("setpci-lnksta", ["setpci", "-s", "1b77:00:00.0", "LnkSta"],
     "PCIe link status via extended config space"),
    ("python3-pass", ["python3", "-c", "pass"], "Python interpreter startup floor"),
]

# Commands needing a specific environment (pcicrawler is installed into an
# isolated target dir; system site-packages is untouched).
PC_ENV = dict(os.environ)
PC_ENV["PYTHONPATH"] = "/tmp/pcicrawler_env"
PC_CMD = ["python3", "-c", "from pcicrawler.cli import main; main()"]
ENVS = {
    "pcicrawler-tree": PC_ENV,
    "pcicrawler-json": PC_ENV,
}
EXTRA_COMMANDS = [
    ("pcicrawler-tree", PC_CMD + ["-t"], "PCIe tree view (isolated pip target dir)"),
    ("pcicrawler-json", PC_CMD + ["-j"], "PCIe JSON view (isolated pip target dir)"),
]


def run_once(argv, env=None):
    t0 = time.perf_counter()
    proc = subprocess.run(argv, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, env=env)
    t1 = time.perf_counter()
    return (t1 - t0) * 1_000_000.0, proc.returncode


def output_sizes(argv, env=None):
    p = subprocess.run(argv, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env)
    return {
        "stdout_bytes": len(p.stdout),
        "stderr_bytes": len(p.stderr),
        "returncode": p.returncode,
    }


def main():
    results = []
    for label, argv, note in COMMANDS + EXTRA_COMMANDS:
        env = ENVS.get(label)
        for _ in range(WARMUP):
            run_once(argv, env)
        samples = []
        rc = None
        for _ in range(RUNS):
            us, rc = run_once(argv, env)
            samples.append(us)
        entry = {
            "label": label,
            "argv": argv,
            "note": note,
            "runs": RUNS,
            "returncode": rc,
            "us_mean": round(statistics.fmean(samples), 1),
            "us_min": round(min(samples), 1),
            "us_max": round(max(samples), 1),
            "us_median": round(statistics.median(samples), 1),
            "us_stdev": round(statistics.pstdev(samples), 1),
            "samples_us": [round(s, 1) for s in samples],
        }
        entry.update(output_sizes(argv, env))
        results.append(entry)
        print(f"{label:32s} mean={entry['us_mean']:>10.1f}us  min={entry['us_min']:>10.1f}  "
              f"max={entry['us_max']:>10.1f}  stdout={entry['stdout_bytes']:>7d}B  rc={rc}", flush=True)

    out = {
        "runs": RUNS,
        "warmup": WARMUP,
        "results": results,
    }
    dest = os.path.join(os.path.dirname(os.path.abspath(__file__)), "raw_measurements.json")
    with open(dest, "w", encoding="utf-8") as fh:
        json.dump(out, fh, indent=2)
    print(f"\nwrote {dest}")


if __name__ == "__main__":
    sys.exit(main())
