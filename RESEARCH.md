# Research: the as-built truth of a server

This document records the market research that produced asbuilt, the decisions it
led to, and the evidence behind them. It was written before the first release and is
kept in the repository so that anyone can check whether the reasoning still holds.

Date of the survey: 2026-11-13. Method: systematic search of the existing tooling,
vendor documentation, kernel sources, public issue trackers and field reports, followed
by a feasibility check of the three candidate niches that survived the first pass.

## The question

Server fleets are built from parts that are individually fine and collectively wrong.
A node can have its PCIe link silently training down to a previous generation, its
memory running one or two speed bins below the rating of the modules, or one socket
populated with a different DIMM kit than the other. None of that is a fault the
operating system reports. The machine boots, the benchmark runs, and the result is
simply lower than the datasheet promised.

The tools that exist answer different halves of that question. `lshw`, `dmidecode`,
`lscpu`, `hwloc`, `lspci` and `turbostat` collect facts. Intel's PerfSpect collects
facts, benchmarks them, and exports them in a dozen formats. Cray's SAT `hwmatch`
compares two machines for hardware mismatches. None of them decides whether the
machine it inspected is correct, and none of them can say what the misconfiguration
costs in bandwidth.

The gap we looked for was therefore not "another hardware inventory tool". It was a way
to turn the as-built state of a server into a verdict, with the performance loss
quantified, in a form that a fleet can be checked against.

## What already exists

The collectors are dense and mostly excellent. `lshw` produces a full device tree and
needs root for completeness (https://manpages.ubuntu.com/manpages/bionic/man1/lshw.1.html).
`dmidecode` is the authoritative SMBIOS dump, and its own man page warns that the
information in the DMI tables is often inaccurate or incomplete
(https://man.archlinux.org/man/dmidecode.8). `lscpu` reads sysfs and `/proc/cpuinfo`
without root and now has machine-readable output (https://manpages.debian.org/testing/util-linux/lscpu.1.en.html).
Intel PCM goes deep on power and bandwidth but requires the `msr` module and root
(https://github.com/intel/pcm). The kernel's own documentation states that the flags in
`/proc/cpuinfo` are "marginally useful for kernel debugging, but not really for
anything else" and points users at `cpuid(1)` instead
(https://kernel.org/doc/html/v6.19-rc4/arch/x86/cpuinfo.html).

Two competitors deserve a closer look. **Intel PerfSpect**
(https://github.com/intel/PerfSpect) is BSD-3 licensed, written in Go, ships a
prebuilt binary, and `perfspect report` already covers BIOS, CPU, memory, DIMM, PCIe,
power, uncore, C-states, frequency, SEL and sensor sections with html, xlsx, json and
txt output, a `--targets` mode for fleets, plus `config` and `benchmark` subcommands.
It is roughly eighty percent of the collecting half of asbuilt. What it does not have
is a verdict layer: no expected-profile comparison, no semantic rules, no envelope
math, and most of its data needs root. **Cray SAT `sat hwmatch`**
(https://cray-hpe.github.io/docs-sat/en-25/introduction/) is MIT licensed and does
compare hardware across a cluster for processor and memory mismatches, which is the
closest thing to fleet drift detection in existence; it is bound to HPE Cray EX and
CSM and covers only CPU and memory. **NVIDIA aicr** (https://github.com/nvidia/aicr,
Apache-2.0) is the closest model to "recipe against actual snapshot", but it is
GPU and Kubernetes scoped.

One important correction to the obvious competitor list: there is no generic open
source `pcie-cli`. The real project with that name is NSA's `paccor` component, an
Electron and Node application for supply chain attestation
(https://nsacyber.github.io/paccor/index.html). The nearest PCIe-focused tool is the
Open Compute Project's `pcicrawler` (https://github.com/opencomputeproject/ocp-diag-pcicrawler),
which is MIT licensed and already emits JSON with capable and current link speed and
width. Its README states that it must run as root, it is a Python package with
dependencies, it emits data rather than verdicts, and it covers PCIe only.

## Candidate one: fleet hardware truth, rules and envelopes — SELECTED

The strongest signal for this niche is that the pain is documented with numbers, in
vendor documents, by the vendors whose hardware it is. Lenovo states that an
unbalanced memory configuration "can reduce the total memory bandwidth to as low as
13% of a balanced memory configuration with 8 identical DIMMs installed per processor"
(https://lenovopress.lenovo.com/lp2000-balanced-memory-configurations-4th-5th-gen-2s-intel-processors).
Dell measured a drop from 65% to 35% of memory bandwidth between unbalanced and balanced
populations with STREAM
(https://dl.dell.com/manuals/all-products/esuprt_software/esuprt_it_ops_datcentr_mgmt/high-computing-solution-resources_White-Papers89_en-us.pdf).
Intel measured a 2.4x geomean speedup simply from disabling NUMA migration
(https://www.intel.com/content/www/us/en/developer/articles/technical/measuring-impact-of-numa-migrations-on-performance.html).
Micron measured 2DPC memory training at 4400 MT/s against 6000 MT/s for 1DPC at equal
capacity, a 27% clock reduction, and a 15% improvement in OLAP query time
(https://www.micron.com/content/dam/micron/global/public/products/memory/dram-modules/rdimm/micron-ddr5-rdimm-higher-perf-scalable-concurrency-tech-brief.pdf).
Dell documents that firmware checks for downgraded PCIe links can miss cases and
directs administrators to the operating system to verify negotiated speed
(https://www.dell.com/support/kbdoc/en-us/000412311/).

That last point is the technical key. The kernel publishes
`current_link_speed`, `current_link_width`, `max_link_speed` and `max_link_width` as
world-readable `DEVICE_ATTR_RO` attributes (0644)
(https://raw.githubusercontent.com/torvalds/linux/master/drivers/pci/pci-sysfs.c), and it
publishes the AER error counters in the same way
(https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-bus-pci-devices-aer).
By contrast, the extended PCI configuration space that `lspci` needs is root-only: the
kernel gives unprivileged callers only the first 64 bytes of configuration space, and
the PCIe capability sits at offset 0x40 and above
(https://raw.githubusercontent.com/torvalds/linux/master/drivers/pci/pci-sysfs.c).
Unprivileged `lspci` therefore cannot see link state at all, which is why the standard
recipe is still `lspci -vvv` plus `grep downgraded` plus root.

What nobody does is compare three numbers against each other: the endpoint's own
capability, the capability of the upstream port above it, and what was actually
negotiated. `lspci` compares a device only with its own capability, which produces
false positives when a device misreports and false negatives when the board caps the
slot. The sysfs attributes make the three-way comparison possible without root, inside
containers, with no dependencies.

The verdict for this candidate is that the niche is genuinely underserved, but that
"MIT plus a single binary" is not differentiation by itself. If asbuilt were only a
collector, PerfSpect would win and the project would die. The defensible part is the
judgement: a rule engine keyed to the platform, an expected-profile schema with a
diff, and a performance-envelope estimate computed from the machine's own numbers.

## Candidate two: PCIe and CXL interconnect validation — PARKED as a product, SELECTED as rules

The interconnect inspection space is crowded: `pciutils` is in every distribution,
`ndctl` carries the CXL utilities with `cxl list` producing JSON for buses, ports,
endpoints, memory devices, decoders and regions plus health, partition and poison
information (https://pmem.io/ndctl/cxl/cxl-list.html). The gap is narrower and real.
The CXL sysfs tree exposes no link speed, no link width, no specification revision and
no per-port link state (https://docs.kernel.org/driver-api/cxl/linux/cxl-driver.html),
so a CXL verdict is not reachable without root. The only widely used recipe remains
shell, `lspci -vv` and `grep`
(https://news.ycombinator.com/item?id=35891290), and vendor runbooks still tell
operators to check `LnkSta` against `LnkCap` by hand
(https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/_sources/troubleshooting/networking_troubleshooting.rst.txt).

Field reports show the cost. A riser that trains down to Gen3 on an otherwise healthy
Gen4 build costs about half the bandwidth and looks fine in every dashboard
(https://kentino.com/nl/pages/gpu-risers-in-ai-server-builds-when-you-need-them-and-what-breaks).
Intel documents erratum SPR132, where a port downgrades after a leaky bucket event and
ring 0 software cannot clear it
(https://edc.intel.com/content/www/us/en/design/products-and-solutions/processors-and-chipsets/eagle-stream/sapphire-rapids-specification-update/017US/errata-details/).
A survey paper on link flapping in the data centre notes that when the equalizer
computation exceeds the LTSSM timeout the link silently falls back to Gen1, a 12.8x
bandwidth reduction with no AER error reported
(https://arxiv.org/pdf/2603.03736v1).

The honest limit is that no authoritative record of a designed topology exists in
sysfs, so "what the board should support" can only be inferred by comparing the
endpoint with the upstream port. That inference is exactly one of the rules asbuilt
implements, and it is worth more inside a whole-machine audit than as a standalone
tool. It stays parked as a separate product.

## Candidate three: CPU capability truth and microarchitecture attribution — KILLED

This candidate was killed, and the reason is that free tools already cover it.
`lscpu -J`, Todd Allen's `cpuid(1)`
(https://manpages.debian.org/unstable/cpuid/cpuid.1.en.gz), the in-tree `kcpuid`,
`libcpuid`, the Rust crate `raw-cpuid`
(https://docs.rs/crate/raw-cpuid/latest), the Go package `klauspost/cpuid`
(https://github.com/klauspost/cpuid) and `rustid`
(https://docs.rs/crate/rustid/1.7.0/source/CHANGELOG.md) all report the instruction
set in tiny binaries, and the kernel itself recommends `kcpuid` and `cpuid(1)` as the
correct answer.

The microarchitecture half is worse for a new entrant. On Ice Lake and later the
kernel exposes top-down metrics directly as fixed counters driven by the
`PERF_METRICS` MSR
(https://kernel.googlesource.com/android/common/+/refs/tags/android14-6.1-2025-09_r1/tools/perf/Documentation/topdown.txt),
and `perf stat --topdown`, `-M TopdownL1` and `-M TopdownL2` produce Retiring, Bad
Speculation, Frontend Bound and Backend Bound with no Python and no per-model event
tables on Intel, and metrics groups on Arm64
(https://learn.arm.com/learning-paths/cross-platform/topdown-compare/2-code-examples/).
Intel's `toplev` remains the deep implementation, but it downloads Intel's per-model
event tables on first run, the SPR core table alone is 376 KB and the uncore
experimental table 1.29 MB
(https://api.github.com/repos/intel/perfmon/contents/SPR/events), and its own issue
tracker is full of complaints about exactly that dependency
(https://github.com/andikleen/pmu-tools/issues/434), about wrong or unstable verdicts
(https://github.com/andikleen/pmu-tools/issues/386,
https://github.com/andikleen/pmu-tools/issues/447) and about nonsense numbers inside
virtual machines (https://github.com/andikleen/pmu-tools/issues/527).

AMD has no equivalent of the fixed-counter path; the closest is the proprietary
AMD uProf pipeline utilisation
(https://docs.amd.com/api/khub/documents/nveV2jevKcum4LePBqpt3w/content), and
`redis-performance/topdown-profiler`
(https://github.com/redis-performance/topdown-profiler) shows that the parts that
matter, uop cache behaviour, memory level parallelism and level two drill-down, need
vendor event sets and precise sampling, which is precisely what a tool without tables
cannot have.

There was one idea worth keeping from this candidate: the fingerprint. Nothing in
existence emits a canonical hash of the ISA feature set so that two nodes can be
compared. asbuilt includes that hash in the manifest, where it belongs, instead of
promising a product.

## What asbuilt is, and what it deliberately is not

asbuilt is a single static binary that reads what the machine actually is, judges it
against the rules that matter for data centre hardware, and prints the verdict with
the evidence and, where the arithmetic allows it, the measured performance loss. It
has no daemon, no database, no configuration file, no telemetry and no GUI. It runs
without root wherever the kernel exposes the data, and it says so, in the manifest,
when a source was unreachable.

Three things are intentionally out of scope. asbuilt is not a benchmark: it does not
measure the workload, because a verdict that depends on the workload is not a verdict
about the hardware. asbuilt is not a monitoring agent: it does not poll, store or
alert, because the same checks run on demand in a CI job and in a provisioning
pipeline without any of that machinery. And asbuilt is not a generic inventory tool:
the manifest exists to be compared, not to be read by a human for hours.

The moat is the rule engine, the expected-profile schema with its diff, and the
envelope arithmetic. The fingerprint is what makes the diff possible. Everything else
is plumbing, and plumbing is supposed to be replaceable.

## Sources

Vendor and standards documents cited above: DMTF SMBIOS Reference Specification 3.9.0
(DSP0134), Intel 600 Series chipset memory speed guidance, Lenovo balanced memory
configuration guidance, Dell PowerEdge downgraded PCIe link knowledge base article,
Dell high computing solution bandwidth white paper, Micron DDR5 RDIMM technical brief,
Intel NUMA migration measurement article, Intel Sapphire Rapids specification update
errata list.

Tools compared: Intel PerfSpect, Cray SAT, NVIDIA aicr, OCP pcicrawler, Intel PCM,
pmu-tools with toplev, OCP pcicrawler, ndctl cxl utilities, pciutils, lshw, dmidecode,
lscpu, hwloc, turbostat, cpuid, kcpuid, libcpuid, raw-cpuid, klauspost/cpuid, rustid,
redis-performance/topdown-profiler, AMD uProf, OpenBMC and Redfish tooling.

Kernel sources: `drivers/pci/pci-sysfs.c` for the link attributes and the
configuration space access rules, the AER sysfs ABI documentation, the CXL driver
documentation, the top-down metrics documentation, and the architecture documentation
on the limits of `/proc/cpuinfo`.
