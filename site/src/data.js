// Data for the asbuilt landing page.
//
// The audit lines are the real output of `asbuilt check` against the
// generated fixture in tests/fixtures/server, not a mock-up.

export const auditLines = [
  { text: "$ asbuilt check", cls: "t-key" },
  { text: "ASBUILT 0.1.0  (manifest schema 1)", cls: "t-dim" },
  { text: "Dell Inc.  PowerEdge R760", cls: "" },
  { text: "BIOS  Dell Inc. 2.10.1 (03/14/2024)", cls: "t-dim" },
  { text: "", cls: "" },
  { text: "MEM   8 channels, 16 DIMMs, 1006644 MB total", cls: "" },
  { text: "      4800 MT/s configured, modules rated 5600 MT/s", cls: "t-signal" },
  { text: "      peak bandwidth 307.2 GB/s (358.4 GB/s at rated speed)", cls: "t-capable" },
  { text: "", cls: "" },
  { text: "PCIE  2 functions, 1 degraded, AER errors on", cls: "" },
  { text: "      0000:01:00.0 16.0 GT/s x16 of 32.0 GT/s x16", cls: "t-signal" },
  { text: "", cls: "" },
  { text: "FINDINGS  2 critical, 3 warn, 2 info", cls: "t-key" },
  { text: "", cls: "" },
  { text: "CRITICAL PCIe link negotiated below endpoint capability", cls: "t-signal" },
  { text: "         impact: pcie_bandwidth -50.0% (31.5 of 63.0 GB/s)", cls: "t-signal" },
  { text: "CRITICAL The processor has been power limit throttled", cls: "t-signal" },
  { text: "         17 power limit throttle events were recorded since boot", cls: "t-dim" },
  { text: "WARN     Memory runs below the rating of the installed modules", cls: "" },
  { text: "         impact: memory_bandwidth -14.3% (307.2 of 358.4 GB/s)", cls: "t-signal" },
];

export const exampleLines = [
  { text: "ASBUILT 0.1.0  (manifest schema 1)", cls: "t-dim" },
  { text: "Dell Inc.  PowerEdge R760", cls: "" },
  { text: "SKU 0KJ3T2  serial CN7792", cls: "t-dim" },
  { text: "BIOS  Dell Inc. 2.10.1 (03/14/2024)", cls: "t-dim" },
  { text: "", cls: "" },
  { text: "CPU   Intel(R) Xeon(R) Platinum 8480+", cls: "" },
  { text: "      2 socket(s) x 28 cores, 2 threads per core, 112 logical CPUs, SMT on", cls: "t-dim" },
  { text: "      2000 MHz base, 3500 MHz max, turbo on, governor performance", cls: "t-dim" },
  { text: "      ISA fingerprint 9472f9464c3f44e4 (78 features)", cls: "t-dim" },
  { text: "      cache L1d 48K, L1i 32K, L2 2048K, L3 107520K", cls: "t-dim" },
  { text: "", cls: "" },
  { text: "MEM   8 channels, 16 DIMMs, 1006644 MB total", cls: "" },
  { text: "      4800 MT/s configured, modules rated 5600 MT/s", cls: "t-signal" },
  { text: "      peak bandwidth 307.2 GB/s (358.4 GB/s at rated speed)", cls: "t-capable" },
  { text: "      ECC on, source smbios", cls: "t-dim" },
  { text: "", cls: "" },
  { text: "NUMA  2 node(s)", cls: "" },
  { text: "", cls: "" },
  { text: "PCIE  2 functions, 1 degraded, AER errors on", cls: "" },
  { text: "      0000:01:00.0 16.0 GT/s x16 of 32.0 GT/s x16", cls: "t-signal" },
  { text: "", cls: "" },
  { text: "OS    Linux version 6.1.0-42-cloud-amd64", cls: "" },
  { text: "      transparent huge pages: always [madvise] never", cls: "t-dim" },
  { text: "      throttling: 0 thermal, 17 power limit events since boot", cls: "t-signal" },
  { text: "", cls: "" },
  { text: "FINDINGS  2 critical, 3 warn, 2 info", cls: "t-key" },
  { text: "", cls: "" },
  { text: "CRITICAL PCIe link negotiated below endpoint capability", cls: "t-signal" },
  { text: "         pcie.link.degraded", cls: "t-dim" },
  { text: "         0000:01:00.0 negotiated 16.0 GT/s x16 while the endpoint supports 32.0 GT/s x16", cls: "" },
  { text: "         impact: pcie_bandwidth -50.0% (31.5 of 63.0 GB/s)", cls: "t-signal" },
  { text: "         evidence: current_link_speed=16.0 GT/s current_link_width=16", cls: "t-dim" },
  { text: "         evidence: max_link_speed=32.0 GT/s max_link_width=16", cls: "t-dim" },
  { text: "         remedy: Seat the card and the riser properly, then check the slot speed setting in firmware.", cls: "" },
  { text: "", cls: "" },
  { text: "CRITICAL The processor has been power limit throttled", cls: "t-signal" },
  { text: "         power.throttle.limit", cls: "t-dim" },
  { text: "         17 power limit throttle events were recorded since boot", cls: "" },
  { text: "         impact: package is spending its budget faster than the budget allows", cls: "t-dim" },
  { text: "", cls: "" },
  { text: "WARN     Memory runs below the rating of the installed modules", cls: "" },
  { text: "         mem.speed.below_module_rating", cls: "t-dim" },
  { text: "         memory runs at 4800 MT/s while the installed modules are rated for 5600 MT/s", cls: "" },
  { text: "         impact: memory_bandwidth -14.3% (307.2 of 358.4 GB/s)", cls: "t-signal" },
];

export const families = [
  {
    id: "pcie",
    name: "PCIe",
    count: 4,
    rules: [
      "Link negotiated below endpoint capability",
      "Endpoint capability above the upstream port limit",
      "Correctable errors were reported",
      "Fatal or non-fatal errors were reported",
    ],
  },
  {
    id: "memory",
    name: "Memory",
    count: 6,
    rules: [
      "Running below the rating of the modules",
      "Mixed rated speeds inside one channel",
      "Mixed part numbers inside one channel",
      "Channels not populated alike",
      "Empty slots",
      "No error correction exposed",
    ],
  },
  {
    id: "numa",
    name: "NUMA",
    count: 2,
    rules: [
      "One node on a multi-socket machine",
      "Nodes holding different amounts of memory",
    ],
  },
  {
    id: "cpu",
    name: "CPU",
    count: 4,
    rules: [
      "Turbo is closed",
      "Governor is not a performance governor",
      "Energy performance bias is not performance",
      "Hardware threads are switched off",
    ],
  },
  {
    id: "firmware",
    name: "Firmware & OS",
    count: 6,
    rules: [
      "Transparent huge pages always on",
      "Automatic NUMA balancing active",
      "PCIe ASPM policy is not performance",
      "A vulnerability without mitigation",
      "Thermal throttling happened",
      "Power limit throttling happened",
    ],
  },
];

export const installSnippets = {
  linux: `# Linux (x86-64)
VERSION=0.1.0
ARCH=amd64
curl -fSLo asbuilt \\
  "https://github.com/Balanov/asbuilt/releases/download/v\${VERSION}/asbuilt-\${VERSION}-linux-\${ARCH}"
mkdir -p "$HOME/.local/bin"
install -m 0755 asbuilt "$HOME/.local/bin/asbuilt"

# verify the checksum, then run
asbuilt --version
asbuilt check`,
  macos: `# macOS (Apple silicon or Intel)
VERSION=0.1.0
ARCH=arm64
curl -fSLo asbuilt \\
  "https://github.com/Balanov/asbuilt/releases/download/v\${VERSION}/asbuilt-\${VERSION}-darwin-\${ARCH}"
mkdir -p "$HOME/.local/bin"
install -m 0755 asbuilt "$HOME/.local/bin/asbuilt"

# remove the quarantine flag for this file only
xattr -d com.apple.quarantine "$HOME/.local/bin/asbuilt"
asbuilt check`,
  windows: `# Windows PowerShell
$Version = "0.1.0"
$Arch = "amd64"
Invoke-WebRequest -Uri "https://github.com/Balanov/asbuilt/releases/download/v$Version/asbuilt-$Version-windows-$Arch.exe" \\
  -OutFile "$env:LOCALAPPDATA\\Programs\\ASBUILT\\asbuilt.exe"

asbuilt.exe --version
asbuilt.exe check`,
  go: `# from source, Go 1.26 or newer
go install github.com/Balanov/asbuilt/cmd/asbuilt@latest

# or build a stripped static binary
git clone https://github.com/Balanov/asbuilt
cd asbuilt
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o asbuilt ./cmd/asbuilt`,
};
