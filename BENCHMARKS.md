# Measurements

Every number on this page was produced by the commands shown, on the machine
described below. Nothing here is an estimate.

## Environment

| | |
|---|---|
| CPU | AMD Ryzen 9 5950X, 16 cores, 32 threads |
| Memory | 16 GB |
| Kernel | 6.18.58.1-microsoft-standard-WSL2, x86_64 |
| Go | 1.27.4 |
| Build | `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w"` |

The timings were taken with the binary and the fixture copied onto the native
Linux filesystem. A build on a Windows-backed mount would measure the mount
rather than the tool.

## Size

| Artifact | Bytes |
|---|---|
| `asbuilt-linux-amd64` | 3,510,432 |
| Go module dependencies | 0 |

The binary is fully static. `ldd` reports `not a dynamic executable`, so it
runs on a server with no runtime, no libc version to match and no package
manager involved. The repository has no third-party Go dependencies at all:
`go.mod` contains nothing but the module line.

## Runtime

The harness is `bench/measure.py` in the repository: 20 timed runs after two
discarded warm-ups, wall clock around `subprocess.run`, stdout and stderr
sent to `DEVNULL`. The audited machine is the fixture in
`tests/fixtures/server`, which describes a dual-socket server with 112
logical CPUs, eight memory channels, sixteen modules and one degraded PCIe
link, so the work per run is the same order as a real audit.

| Operation | Mean | Min | Max |
|---|---:|---:|---:|
| Full audit: manifest, 33 rules, text report | 3,158 µs | 2,716 µs | 3,706 µs |
| Canonical JSON manifest | 3,290 µs | 2,945 µs | 3,993 µs |
| JSON manifest with `--json` | 3,380 µs | 2,852 µs | 4,163 µs |
| `--version` (process startup floor) | 1,753 µs | 1,535 µs | 2,148 µs |
| Live audit of the running machine | 4,485 µs | 3,961 µs | 5,343 µs |

The interesting number is the last pair. The floor of about 1.75 ms is what
fork, exec and the Go runtime cost before asbuilt does anything; a full audit
of a 112-thread server with a firmware table adds roughly 1.4 ms of actual
work. On a real bare-metal server the startup floor is lower, because the
kernel does not have to cross a Windows host boundary.

## Against the other tools

Same harness, same machine, same commands. The full dataset, including the
caveats, is in [bench/COMPETITORS.md](bench/COMPETITORS.md).

| Tool | Command | Mean | Root | Verdict |
|---|---|---:|:---:|---|
| **asbuilt** | `--root … check` | **3,158 µs** | no | full audit, 33 rules |
| lshw | `-json` | 136,686 µs | yes | tree output only |
| pcicrawler | `-j` | 50,852 µs | yes | PCIe only |
| lspci | `-vvv` | 12,632 µs | yes | two numbers compared |
| lscpu | `-J` | 8,241 µs | no | CPU only |
| dmidecode | — | 686 µs* | yes | SMBIOS dump, no verdict |

\* `dmidecode`'s 0.69 ms is **not** a real DMI dump: this WSL kernel exposes
no SMBIOS table, so the tool prints "No SMBIOS nor DMI entry point found" and
exits 0. Its timing measures failing to find a table, not reading one.

That makes asbuilt **43x faster than lshw** for a complete machine audit, and
the only tool in the set that produces a verdict at all: the others collect
facts, and asbuilt is the one that decides what they mean.

## Reproducing

```sh
go build -trimpath -ldflags "-s -w" -o asbuilt ./cmd/asbuilt
ls -l asbuilt

# the harness, on a Linux host
python3 bench/measure.py
```

## Coverage of the fixture

The fixture in `tests/fixtures/server` is generated, not recorded, so that it
stays correct when the code changes:

```sh
go run ./tools/fixturegen -out tests/fixtures/server
```

It reproduces the sources asbuilt reads: `/proc/cpuinfo`, `/proc/meminfo`,
`/proc/version`, `/proc/uptime`, `/proc/cmdline`, the cpufreq and cache trees
under `/sys/devices/system/cpu`, the NUMA nodes, the thermal throttle
counters, the PCIe devices with their link attributes and AER counters, and a
SMBIOS table describing a Dell PowerEdge R760 with eight DDR5 channels.

Because the fixture is a directory tree, it also proves that asbuilt can audit
a machine it does not own: `--root` points every read at a different tree,
which is how the same code path serves a container image, a chroot and a test.

## What the fixture demonstrates

The fixture is deliberately wrong in four ways, and asbuilt finds all of them:

| Finding | Severity | Measured loss |
|---|---|---|
| PCIe link negotiated Gen4 x16 instead of Gen5 x16 | critical | 50% of the link bandwidth, 31.5 of 63.0 GB/s |
| Memory running at 4800 MT/s while the modules are rated for 5600 MT/s | warn | 14.3% of the memory bandwidth, 307.2 of 358.4 GB/s |
| 17 power limit throttle events since boot | critical | not quantified: the counter says it happened, not how much it cost |
| 42 correctable PCIe errors on the same link | warn | a rising rate means the link is losing margin |

The two quantified losses are computed from the machine's own numbers: the
link bandwidth from the negotiated speed and width with the right line coding,
the memory bandwidth from the number of populated channels and the configured
data rate.
