// Data for the asbuilt landing page.
//
// The audit lines are the real output of `asbuilt check` against the
// generated fixture in tests/fixtures/server, not a mock-up.

export const auditLines = [
  { text: "$ asbuilt check", cls: "t-key" },
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

export const families = [
  { id: "pcie", count: 4 },
  { id: "memory", count: 6 },
  { id: "numa", count: 2 },
  { id: "cpu", count: 4 },
  { id: "firmware", count: 6 },
];

export const installSnippets = {
  linux: `# Linux (x86-64)
VERSION=0.1.0
ARCH=amd64
curl -fSLo asbuilt \\
  "https://github.com/Balanov/asbuilt/releases/download/v\${VERSION}/asbuilt-\${VERSION}-linux-\${ARCH}"
mkdir -p "$HOME/.local/bin"
install -m 0755 asbuilt "$HOME/.local/bin/asbuilt"
asbuilt --version`,
  macos: `# macOS (Apple silicon or Intel)
VERSION=0.1.0
ARCH=arm64
curl -fSLo asbuilt \\
  "https://github.com/Balanov/asbuilt/releases/download/v\${VERSION}/asbuilt-\${VERSION}-darwin-\${ARCH}"
mkdir -p "$HOME/.local/bin"
install -m 0755 asbuilt "$HOME/.local/bin/asbuilt"
xattr -d com.apple.quarantine "$HOME/.local/bin/asbuilt"
asbuilt --version`,
  windows: `# Windows PowerShell
$Version = "0.1.0"
$Arch = "amd64"
Invoke-WebRequest -Uri "https://github.com/Balanov/asbuilt/releases/download/v$Version/asbuilt-$Version-windows-$Arch.exe" \\
  -OutFile "$env:LOCALAPPDATA\\Programs\\ASBUILT\\asbuilt.exe"
asbuilt.exe --version`,
  go: `# from source, Go 1.26 or newer
go install github.com/Balanov/asbuilt/cmd/asbuilt@latest

# or a stripped static binary
git clone https://github.com/Balanov/asbuilt
cd asbuilt
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o asbuilt ./cmd/asbuilt`,
};
