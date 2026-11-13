<div align="center">
  <img src="assets/brand/logo-512.png" alt="asbuilt — the as-built truth of a server" width="128">
  <p><strong>Know what the machine actually is, and what that costs you.</strong></p>
  <p>Linux · x86_64 · static binary · no root · no dependencies · MIT</p>
  <p>
    <a href="#quick-start">Quick start</a> ·
    <a href="#what-it-checks">What it checks</a> ·
    <a href="#profiles">Profiles</a> ·
    <a href="#manifests-and-diffs">Manifests</a> ·
    <a href="#ci-gate">CI gate</a> ·
    <a href="BENCHMARKS.md">Measurements</a> ·
    <a href="RESEARCH.md">Research</a> ·
    <a href="README_RU.md">Русский</a>
  </p>
  <p>
    <a href="https://github.com/Balanov/asbuilt/actions/workflows/ci.yml"><img src="https://github.com/Balanov/asbuilt/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="https://github.com/Balanov/asbuilt/releases/latest"><img src="https://img.shields.io/github/v/release/Balanov/asbuilt" alt="Latest release"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg" alt="MIT license"></a>
    <a href="go.mod"><img src="https://img.shields.io/badge/deps-0-success.svg" alt="No dependencies"></a>
  </p>
</div>

ASBUILT is the as-built truth of a server. It reads what the machine actually
is, judges it against the rules that matter for data centre hardware, and prints
the verdict with the evidence and, where the arithmetic allows it, the measured
performance loss.

A server can be built from parts that are individually fine and collectively
wrong. Its PCIe link trains down to a previous generation. Its memory runs a
speed bin below the rating of the modules. One socket is populated with a
different kit than the other. The machine boots, the benchmark runs, and the
result is simply lower than the datasheet promised. No operating system reports
any of it.

```console
$ asbuilt check
ASBUILT 0.1.0  (manifest schema 1)
Dell Inc.  PowerEdge R760
SKU 0KJ3T2  serial CN7792
BIOS  Dell Inc. 2.10.1 (03/14/2024)
BOARD Dell Inc. 0KJ3T2

CPU   Intel(R) Xeon(R) Platinum 8480+
      2 socket(s) x 28 cores, 2 threads per core, 112 logical CPUs, SMT on
      2000 MHz base, 3500 MHz max, turbo on, governor performance
      ISA fingerprint 9472f9464c3f44e4 (78 features)
      cache L1d 48K, L1i 32K, L2 2048K, L3 107520K

MEM   8 channels, 16 DIMMs, 1006644 MB total
      4800 MT/s configured, modules rated 5600 MT/s
      peak bandwidth 307.2 GB/s (358.4 GB/s at rated speed)
      ECC on, source smbios

NUMA  2 node(s)

PCIE  2 functions, 1 degraded, AER errors on
      0000:01:00.0 16.0 GT/s x16 of 32.0 GT/s x16

OS    Linux version 6.1.0-42-cloud-amd64 ...
      transparent huge pages: always [madvise] never
      throttling: 0 thermal, 17 power limit events since boot

FINDINGS  2 critical, 3 warn, 2 info

CRITICAL PCIe link negotiated below endpoint capability
         pcie.link.degraded
         0000:01:00.0 negotiated 16.0 GT/s x16 while the endpoint supports 32.0 GT/s x16
         impact: pcie_bandwidth -50.0% (31.5 of 63.0 GB/s)
         evidence: current_link_speed=16.0 GT/s current_link_width=16
         evidence: max_link_speed=32.0 GT/s max_link_width=16
         remedy: Seat the card and the riser properly, then check the slot speed setting in firmware.
```

## Why asbuilt

- **Judgement, not inventory.** `lshw`, `dmidecode`, `lscpu` and `lspci` collect
  facts; Intel PerfSpect collects them very well. None of them decides whether
  the machine is correct, and none of them can say what the misconfiguration
  costs. That layer is the product.
- **Quantified losses.** A finding says how much bandwidth is missing, computed
  from the machine's own numbers: 50% of the link, 14.3% of the memory. A
  finding that cannot be quantified says so instead of inventing a number.
- **Root-optional.** The kernel publishes the negotiated PCIe link state and the
  AER counters as world-readable sysfs attributes, so the degraded link check
  works unprivileged, inside containers, without `pciutils`. Root adds the
  firmware table, which is where the memory population lives.
- **Explicit about what it could not read.** A missing source becomes a note in
  the manifest, never a silent omission.
- **No daemon, no database, no telemetry, no GUI.** One static binary that
  prints to stdout and exits with a code.

## Install

### Linux

```sh
VERSION=0.1.0
ARCH=amd64  # Use arm64 on ARM64.
curl -fSLo asbuilt "https://github.com/Balanov/asbuilt/releases/download/v${VERSION}/asbuilt-linux-${ARCH}"
mkdir -p "$HOME/.local/bin"
install -m 0755 asbuilt "$HOME/.local/bin/asbuilt"
export PATH="$HOME/.local/bin:$PATH"
asbuilt --version
```

Make sure `$HOME/.local/bin` is on your `PATH`. Every release includes
`checksums.txt`:

```sh
curl -fSLO "https://github.com/Balanov/asbuilt/releases/download/v${VERSION}/checksums.txt"
grep "asbuilt-linux-${ARCH}$" checksums.txt | sha256sum -c -
```

The binaries are not code-signed yet, so SmartScreen or Gatekeeper may warn.
Verify the SHA-256 checksum before allowing a downloaded executable. On macOS,
after verification, remove quarantine from this file only:

```sh
xattr -d com.apple.quarantine "$HOME/.local/bin/asbuilt"
```

### Build from source

With Go 1.26 or newer:

```sh
go install github.com/Balanov/asbuilt/cmd/asbuilt@latest
```

or, in a clone:

```sh
go build -trimpath -ldflags "-s -w" -o asbuilt ./cmd/asbuilt
```

## Quick start

Try it without a server and without root. The repository ships a generated
fixture that reproduces a misconfigured dual-socket server:

```sh
go run ./tools/fixturegen -out tests/fixtures/server
asbuilt --root tests/fixtures/server check
```

Then audit a real machine:

```sh
asbuilt check
asbuilt check --json
asbuilt manifest > node-014.json
```

## What it checks

Thirty-three rules in five families. Every rule carries the evidence it used
and a source that justifies it, and `asbuilt explain` prints both.

| Family | Rules |
|---|---|
| PCIe | link negotiated below the endpoint capability, endpoint capability above the upstream port limit, correctable errors, fatal and non-fatal errors |
| Memory | running below the rating of the modules, mixed rated speeds in a channel, mixed part numbers in a channel, channels not populated alike, empty slots, no error correction |
| NUMA | one node on a multi-socket machine, nodes holding different amounts of memory |
| CPU | turbo closed, governor not a performance governor, energy performance bias, SMT disabled |
| Firmware and OS | transparent huge pages always on, automatic NUMA balancing, PCIe ASPM policy, a vulnerability without mitigation, thermal throttling, power limit throttling |

The PCIe comparison is the part that no other tool does. `lspci` compares a
device with its own capability, which misses the case where the board or the
firmware capped the slot, and flags the case where a device misreports. asbuilt
compares three numbers: what the endpoint can do, what the upstream port
allows, and what was actually negotiated.

```console
$ asbuilt rules
RULES  33

CRITICAL pcie.link.degraded                   PCIe link negotiated below endpoint capability
CRITICAL mem.population.uneven                Memory channels are not populated alike
...
```

## Profiles

A profile is how the fleet is supposed to be built, in plain JSON, so that a
hardware team can keep it next to the inventory.

```json
{
  "name": "xd-node",
  "expect": {
    "platform": { "product": "PowerEdge R760" },
    "cpu": { "sockets": 2, "threads_per_core": 2, "turbo_enabled": true },
    "memory": { "channels": 8, "dimms_per_channel": 2, "configured_mts": 5600, "ecc_enabled": true },
    "pcie": { "min_link_gts": 32 },
    "os": { "transparent_hugepage": "madvise", "governor": "performance" }
  }
}
```

```console
$ asbuilt --profile profiles/xd-node.json check
...
CRITICAL Memory runs at a speed other than the expected profile
         profile.memory.configured_mts
         memory runs at 4800 MT/s, expected 5600 MT/s
         impact: memory_bandwidth -14.3% (307.2 of 358.4 GB/s)
```

## Manifests and diffs

`asbuilt manifest` prints a canonical JSON manifest: the firmware identity, the
processor, the memory population per channel, the PCIe links, the operating
system state, and a fingerprint of the instruction set the kernel exposes.

```sh
asbuilt manifest > node-014.json
asbuilt diff node-014.json node-015.json
```

```console
$ asbuilt diff node-014.json node-015.json
DIFF  3 difference(s)

  memory.configured_mts
    a: 5600
    b: 4800
    affects: memory bandwidth

  pcie.0000:01:00.0.current_speed_gts
    a: 32.0
    b: 16.0
    affects: pcie bandwidth
```

The diff is what turns a pile of JSON into an audit: two nodes built the same
way should produce an empty diff, and every line that appears is either a
configuration mistake or a hardware defect.

## CI gate

`asbuilt` exits non-zero when a finding reaches the threshold, so a
provisioning pipeline can refuse a node that is not what was ordered.

```sh
asbuilt check --fail-on critical || exit 1
```

| Exit code | Meaning |
|---|---|
| 0 | no finding reached the threshold |
| 1 | a finding reached the threshold |
| 2 | the audit could not run |

## Design notes

The rule engine is separated from the collectors on purpose. The collectors
read the kernel and the firmware; the engine decides what those facts mean.
That split is what makes the tool forkable: a hardware team that knows its
platform can add rules without touching the parsers, and a parser bug can be
fixed without changing a verdict.

Every collector is read-only and works against a root-relative filesystem, so
the same code path serves a live server, a container image and a recorded
fixture tree. The SMBIOS decoder follows the DMTF specification 3.9.0 and was
written against it rather than against a sample dump; the memory population,
the configured speed and the rated speed are three different fields, and the
distinction between them is the whole point of the memory rules.

## Documentation

- [RESEARCH.md](RESEARCH.md) — the market research, the three candidates that
  were considered, and why this one was built.
- [BENCHMARKS.md](BENCHMARKS.md) — measured size, runtime and coverage.
- [CHANGELOG.md](CHANGELOG.md) — what changed.
- `asbuilt explain <rule-id>` — the rationale and the source for any rule.

## License

MIT. See [LICENSE](LICENSE). Fork it, change it, ship it.
