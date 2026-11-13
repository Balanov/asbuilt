# asbuilt vs. existing Linux hardware-audit tools

**Measured on:** WSL2 Ubuntu 24.04.5 (WSL distro `UbuntuASB`), kernel `6.18.58.1-microsoft-standard-WSL2`, 32 logical CPUs (AMD Ryzen 9 5950X), 2026-11-13.
**Machine-readable twin:** [`competitors.json`](competitors.json). **Raw samples:** [`raw_measurements.json`](raw_measurements.json).

Every number below was measured on that one machine, or is marked **sourced** with the upstream URL it came from. Where a tool could not be run here, the cells say so instead of estimating.

## Headline

| | asbuilt | lshw | pcicrawler | lscpu | lspci | dmidecode |
|---|---|---|---|---|---|---|
| Whole-machine audit | **3.16 ms** | 136.7 ms (43× slower) | 50.9 ms (PCI only) | 8.2 ms (CPU only) | 12.6 ms (PCI only) | 0.69 ms\* |
| Verdicts? | yes (33 rules) | no | no | no | no | no |
| Needs root | **no** | yes | yes | no | yes (for extended config) | yes |

\* `dmidecode`'s 0.69 ms is **not** a real DMI dump: this WSL kernel exposes no SMBIOS table, so it prints "No SMBIOS nor DMI entry point found, sorry." and exits 0. Its timing measures *failing to find* a table, not reading one.

## Timings (mean of 20 runs, microseconds)

`runs` = 20 for every measured row; 2 warm-up runs discarded.

| Tool | Command | mean µs | min µs | max µs | needs root | install bytes | output bytes | Status |
|---|---|---:|---:|---:|:---:|---:|---:|---|
| **asbuilt** | `--root /tmp/server check` | **3 158** | 2 716 | 3 706 | no | 3 510 432 | 4 868 | Measured |
| asbuilt (canonical JSON) | `--root /tmp/server manifest` | 3 290 | 2 945 | 3 993 | no | 3 510 432 | 11 299 | Measured |
| asbuilt (JSON with `--json`) | `--root /tmp/server check --json` | 3 380 | 2 852 | 4 163 | no | 3 510 432 | 11 299 | Measured |
| asbuilt (startup floor) | `--version` | 1 753 | 1 535 | 2 148 | no | 3 510 432 | 14 | Measured |
| asbuilt (live host) | `check` | 4 485 | 3 961 | 5 343 | no | 3 510 432 | 2 167 | Measured |
| **lshw** | `-json` | 136 686 | 132 506 | 151 208 | yes | 997 376 | 13 638 | Measured |
| lshw | `-short` | 140 726 | 133 229 | 155 126 | yes | 997 376 | 853 | Measured |
| **lscpu** | `-J` | 8 241 | 7 855 | 8 623 | no | 3 521 536 | 4 581 | Measured |
| **lspci** | `-vvv` | 12 632 | 11 811 | 15 994 | yes | 186 368 | 3 102 | Measured |
| setpci | `-s 1b77:00:00.0 LnkSta` | 860 | 712 | 1 283 | yes | 186 368 | 0 (rc=1) | Measured (fails) |
| **dmidecode** | (no args) | 686 | 591 | 820 | yes | 233 472 | 97 | Measured (no DMI table) |
| dmidecode | `-t 17` | 676 | 614 | 809 | yes | 233 472 | 97 | Measured (no DMI table) |
| **pcicrawler** | `-j` | 50 852 | 46 531 | 64 534 | yes | 109 365 | 483 | Measured |
| pcicrawler | `-t` | 47 706 | 44 812 | 51 197 | yes | 109 365 | 153 | Measured |
| python3 | `-c pass` | 8 982 | 8 288 | 10 512 | no | 82 944 | 0 | Measured (reference) |
| **PerfSpect** | `perfspect report` | — | — | — | yes | 45 663 339 | — | **Sourced** — not runnable here |
| **sat hwmatch** | `sat hwmatch` | — | — | — | yes | — | — | **Sourced** — not runnable here |

Install sizes: `dpkg Installed-Size` × 1024 for apt packages; `stat` size for the asbuilt binary; byte size of pcicrawler's own files in an isolated `pip --target` dir (excluding `click`/`six`); the released `perfspect.tgz` asset for PerfSpect.

## How it was measured

Timing harness `bench/measure.py` wraps each command in `subprocess.run()` and measures wall time with Python 3.12's `time.perf_counter()`; stdout and stderr go to `DEVNULL` so the pipe is never measured. Each command runs **20 timed iterations after 2 discarded warm-up runs**, and all commands execute on the same WSL2 instance. The asbuilt binary and the machine fixture were copied to the **native Linux filesystem** (`/tmp/asbuilt`, `/tmp/server`) first — running them from the `/mnt/f` Windows-backed mount would have measured the mount instead of the tool. `install_size_bytes` comes from `dpkg-query -W -f='${Installed-Size}'`, from `stat`, or from the isolated `pip --target` directory, as noted per row. The fixture (`tests/fixtures/server`) describes a dual-socket Dell PowerEdge R760 with 112 logical CPUs, 8 memory channels and one degraded PCIe link; **that description belongs to the fixture, not to the measuring host** (32 logical CPUs, AMD Ryzen 9 5950X). The `asbuilt` row audits the fixture; `asbuilt (live host)` audits the real machine.

### Two caveats that matter for reading this table

1. **asbuilt accepts `--json` before or after the subcommand.** `asbuilt --root /tmp/server manifest` prints the canonical JSON manifest (11 299 B), and `asbuilt --root /tmp/server check --json` prints the manifest together with the findings. An earlier revision of this dataset reported the opposite: the subcommand parser used to stop at the first non-flag argument, so a flag placed *before* the subcommand was silently ignored. That was a real bug, it is fixed, and the numbers above were re-measured afterwards.
2. **The three virtio devices on this WSL host are not PCIe endpoints.** `setpci` 3.10.0 has no `LnkSta` named register (it exposes `ECAP_<capability>` names only), `/sys/bus/pci/devices/*/current_link_speed` does not exist on this kernel, and no real PCIe endpoint is present — so **PCIe link status could not be measured here at all**. asbuilt's link-degradation verdict comes from the fixture, which carries the synthetic `current_link_speed`/`max_link_speed` attributes.

## Capability comparison

| Capability | asbuilt | Competitors |
|---|---|---|
| Whole-machine audit, one invocation | 3.08 ms, manifest + 33 rules + report + verdicts | lshw 227 ms (tree only, no rules); pcicrawler 57 ms (PCI only); lscpu 7.8 ms (CPU only); lspci 11.5 ms (PCI only) |
| DIMM population (slot fill, part numbers, rated vs configured speed) | 16 DIMMs / 8 channels, per-module rated speed, `mem.parts.mixed_within_channel`, `mem.speed.below_module_rating` | lscpu: **impossible** — 0 `dimm`/`slot`/`module` keys in `lscpu -J`; dmidecode `-t 17`: the only equivalent, unusable on this kernel; lshw: aggregate "System memory" only |
| PCIe link degradation verdict | `pcie.link.degraded` CRITICAL, negotiated 16.0 GT/s x16 vs 32.0 GT/s x16 capability, **impact: pcie_bandwidth −50.0%** (31.5 of 63.0 GB/s), with evidence + remedy | lspci/setpci: no verdict and **not measurable here**; pcicrawler: exposes capable/cur/target speed + widths + AER, but no pass/fail judgement |
| Expected-profile diff | `asbuilt diff A.json B.json` and `--profile FILE` with `profile.*` rules (`profile.memory.ecc`, `profile.pcie.min_link`, `profile.platform.mismatch`, …) | PerfSpect: **no compare/expected-profile command** — documented commands are report, benchmark, metrics, telemetry, flamegraph, lock, config, update, extract; global flags are `--debug/--log-stdout/--noupdate/--output/--syslog/--tempdir` only |
| Semantic performance-loss verdicts | quantified impact lines: `pcie_bandwidth -50.0%`, `memory_bandwidth -14.3%` | PerfSpect reports raw/threshold-flagged values: v3.10.0 notes the metrics HTML report "highlights metrics that exceed a threshold, and provides context and/or a tip"; the benchmark command emits measured speed/power/temperature numbers, not loss verdicts |
| CPU/memory uniformity across sockets/cards/slots | `mem.population.uneven`, `mem.parts.mixed_within_channel`, `mem.speed.mixed_within_channel`, `numa.asymmetric`, `numa.single_node_multi_socket` — on any Linux server | **sat hwmatch** is the closest analogue but is Cray-only and CPU/memory-only: "Processor values that must match are model, core count, and speed. Memory module values that must match are memory type, memory device type, and size. The number of memory modules must also match." — matched at slot/card/node level, requiring an HPE Cray EX (Shasta) system + CSM Hardware State Manager + API-gateway auth |
| Root required for a complete answer | **no** — verified identical 4 868 B output and rc=0 as user `nobody`, fixture and live | lshw: warns "WARNING: you should run this program as super-user", output incomplete; dmidecode: needs `/dev/mem` (`crw------- root root`), rc=1 without root; lspci: `Capabilities: <access denied>` without root; pcicrawler: "Must run as root as it uses privileged sysfs entries"; PerfSpect `report`: no `--noroot` flag (only `metrics` has one); sat hwmatch: root/crayadm on a K8s manager node |
| Configuration breadth in one report | 33 rules across CPU, memory, NUMA, PCIe, OS/power/virtualization + canonical JSON manifest (schema 1) | PerfSpect `report` is broader — 33 configuration categories (bios, os, software, cpu, prefetcher, isa, accelerator, power, cstates, frequency, sst, uncore, elc, memory, dimm, netconfig, nic, disk, filesystem, gpu, gaudi, cxl, pcie, cve, process, sensor, chassisstatus, pmu, sel, kernellog) — but it **reports** configuration, it does not judge it |
| PCIe topology with speed/width per link | `PCIE` report section: negotiated vs capability speed/width, AER state, degraded-link finding | pcicrawler `-t`/`-j`: dedicated topology tree with capable/cur/target speed, widths, AER (`--aer`) and VPD (`--vpd`), filterable by class/device/address — the best open-source PCIe specialist here, still no verdict layer |
| License (redistribution cost) | MIT | lshw GPL-2+; dmidecode GPL-2+; pciutils GPL-2+; util-linux/lscpu GPL-2+; PerfSpect BSD-3-Clause; pcicrawler MIT; Cray SAT CLI MIT (SAT product itself HPE-proprietary) |

## Licenses — checked against official sources

| Tool | License | Evidence |
|---|---|---|
| asbuilt | MIT | local `F:\GitHub\asbuilt\LICENSE` (MIT, Copyright (c) 2026 Anton Balanov) |
| lshw | GPL-2+ | [`hardware-probe/lshw`](https://github.com/hardware-probe/lshw); `/usr/share/doc/lshw/copyright`: "GNU General Public License … either version 2, or (at your option) any later version" |
| dmidecode | GPL-2+ | [`nmhughes/dmidecode`](https://github.com/nmhughes/dmidecode); `/usr/share/doc/dmidecode/copyright`: `License: GPL-2+` |
| pciutils (lspci/setpci) | GPL-2+ | [`gitlab.com/pciutils/pciutils`](https://gitlab.com/pciutils/pciutils); `/usr/share/doc/pciutils/copyright`: `License: GPL-2+` |
| util-linux (lscpu) | GPL-2+ | [`kernel.org/…/util-linux`](https://www.kernel.org/pub/linux/utils/util-linux/); `/usr/share/doc/util-linux/copyright`: `License: GPL-2+` |
| Intel PerfSpect | BSD-3-Clause | [`intel/PerfSpect/blob/main/LICENSE`](https://github.com/intel/PerfSpect/blob/main/LICENSE) — "BSD 3-Clause License, Copyright (C) 2024 Intel Corporation" |
| OCP pcicrawler | MIT | [`ocp-diag-pcicrawler/blob/main/LICENSE`](https://github.com/opencomputeproject/ocp-diag-pcicrawler/blob/main/LICENSE) — "MIT License, Copyright (c) Meta Platforms, Inc. and affiliates." |
| Cray SAT (`sat` CLI) | MIT for the CLI source; the SAT product is HPE-proprietary | [`Cray-HPE/sat/blob/main/LICENSE`](https://github.com/Cray-HPE/sat/blob/main/LICENSE) — "MIT License, (C) Copyright 2019-2022 Hewlett Packard Enterprise Development LP" |

## Not measurable here — and why

| Tool | Why it could not be measured | What is still known (sourced) |
|---|---|---|
| **Intel PerfSpect** | Requires an Intel Xeon platform with PMU/perfmon access. This host is an AMD Ryzen 9 on WSL2, which exposes no perfmon access, so `perfspect metrics`/`report` cannot produce data. | BSD-3-Clause; `perfspect.tgz` for v3.18.0 is **45 663 339 B** (published 2026-11-06, `oss_source.tgz` 61 987 112 B); `report` covers 33 categories; `--noroot` exists only on `metrics`; no expected-profile/diff command and no semantic loss verdicts are documented. |
| **Cray SAT `sat hwmatch`** | `sat` runs inside a Podman container on HPE Cray EX Kubernetes manager nodes and queries CSM's Hardware State Manager; it needs the whole HPE Cray EX (Shasta) stack plus API-gateway authentication. It is not installable on a generic Linux host. | MIT CLI; scope is CPU/memory only, matched at slot/card/node level; requires root/crayadm on an `ncn-m` node. |
| **PCIe link status via `setpci`/`lspci`** | No `LnkSta` register in pciutils 3.10.0 (`ECAP_<capability>` names only), no `current_link_speed` sysfs attribute on this kernel, and no real PCIe endpoint on the host — only 3 virtio devices. | `lspci -vvv` without root prints `Capabilities: <access denied>`, so extended config space does need root. |

## Reproducing

```bash
# one-time setup inside WSL
wsl -d UbuntuASB -e /bin/bash -lc "apt-get update && apt-get install -y --no-install-recommends lshw dmidecode pciutils python3 python3-pip"
wsl -d UbuntuASB -e /bin/bash -lc "cp /mnt/f/GitHub/asbuilt/bin/asbuilt-linux-amd64 /tmp/asbuilt && chmod +x /tmp/asbuilt && cp -r /mnt/f/GitHub/asbuilt/tests/fixtures/server /tmp/server"

# run the benchmark (writes bench/raw_measurements.json)
wsl -d UbuntuASB -e /bin/bash -lc "cp /mnt/f/GitHub/asbuilt/bench/measure.py /tmp/measure.py && python3 /tmp/measure.py"

# validate the dataset
wsl -d UbuntuASB -e /bin/bash -lc "python3 /mnt/f/GitHub/asbuilt/bench/validate.py"
```

`BENCH_RUNS` and `BENCH_WARMUP` override the 20/2 defaults. pcicrawler is installed into an isolated `pip --target /tmp/pcicrawler_env` directory (system site-packages untouched) and invoked via `PYTHONPATH`.

## Sources

All URLs re-fetched and checked on **2026-11-13**.

- Intel PerfSpect README — <https://github.com/intel/PerfSpect/blob/main/README.md> (BSD-3 badge; "Metrics Without Root Permissions" section; remote-target root note)
- Intel PerfSpect global options — <https://github.com/intel/PerfSpect/blob/main/docs/perfspect.md>
- Intel PerfSpect `report` options (33 categories) — <https://github.com/intel/PerfSpect/blob/main/docs/perfspect_report.md>
- Intel PerfSpect `metrics` options (`--noroot`) — <https://github.com/intel/PerfSpect/blob/main/docs/perfspect_metrics.md>
- Intel PerfSpect releases — <https://github.com/intel/PerfSpect/releases> ; v3.18.0 assets/sizes — <https://api.github.com/repos/intel/PerfSpect/releases/latest> ; v3.10.0 notes (threshold highlighting) — <https://github.com/intel/PerfSpect/releases/tag/v3.10.0>
- Intel PerfSpect LICENSE — <https://github.com/intel/PerfSpect/blob/main/LICENSE>
- OCP pcicrawler README ("Must run as root as it uses privileged sysfs entries") — <https://github.com/opencomputeproject/ocp-diag-pcicrawler>
- OCP pcicrawler setup.py (`install_requires = ["click", "six"]`, `python_requires >= 3.6.0`, `license="MIT"`) — <https://github.com/opencomputeproject/ocp-diag-pcicrawler/blob/main/setup.py>
- OCP pcicrawler LICENSE — <https://github.com/opencomputeproject/ocp-diag-pcicrawler/blob/main/LICENSE>
- Cray SAT introduction (`sat hwmatch` → "Report hardware mismatches", depends on CSM Hardware State Manager) — <https://cray-hpe.github.io/docs-sat/en-25/introduction/>
- Cray SAT `sat-hwmatch` man page (exact CPU/memory field scope, slot/card/node levels) — <https://github.com/Cray-HPE/sat/blob/main/docs/man/sat-hwmatch.8.rst>
- Cray SAT LICENSE — <https://github.com/Cray-HPE/sat/blob/main/LICENSE>
- lshw — <https://github.com/hardware-probe/lshw>
- dmidecode — <https://github.com/nmhughes/dmidecode>
- pciutils — <https://gitlab.com/pciutils/pciutils>
- util-linux — <https://www.kernel.org/pub/linux/utils/util-linux/>
- Local package licence texts (as installed): `/usr/share/doc/{lshw,dmidecode,pciutils,util-linux}/copyright`
