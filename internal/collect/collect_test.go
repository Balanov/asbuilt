package collect

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Balanov/asbuilt/internal/dmitest"
	"github.com/Balanov/asbuilt/internal/sysfs"
)

// fixture writes a file into a tree, creating directories as needed.
func fixture(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(path, "/")))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
}

const cpuInfoTwoSockets = `processor	: 0
vendor_id	: GenuineIntel
cpu family	: 6
model		: 143
model name	: Intel(R) Xeon(R) Platinum 8480+
stepping	: 8
microcode	: 0x2b000603
cpu MHz		: 2000.000
cache size	: 107520 KB
physical id	: 0
siblings	: 56
core id		: 0
cpu cores	: 28
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology cpuid tsc_known_freq pni pclmulqdq ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt aes xsave avx f16c rdrand hypervisor lahf_lm abm 3dnowprefetch invpcid_single ssbd ibrs ibpb stibp fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid mpx rdseed adx smap clflushopt xsaveopt xsavec xgetbv1 arat
bugs		:
bogomips	: 4000.00
TLB size	: 0
address sizes	: 46 bits physical, 57 bits virtual
power management:

processor	: 1
vendor_id	: GenuineIntel
cpu family	: 6
model		: 143
model name	: Intel(R) Xeon(R) Platinum 8480+
stepping	: 8
microcode	: 0x2b000603
cpu MHz		: 2000.000
cache size	: 107520 KB
physical id	: 1
siblings	: 56
core id		: 0
cpu cores	: 28
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology cpuid tsc_known_freq pni pclmulqdq ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt aes xsave avx f16c rdrand hypervisor lahf_lm abm 3dnowprefetch invpcid_single ssbd ibrs ibpb stibp fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid mpx rdseed adx smap clflushopt xsaveopt xsavec xgetbv1 arat
bugs		:
bogomips	: 4000.00
TLB size	: 0
address sizes	: 46 bits physical, 57 bits virtual
power management:
`

// buildFixture writes a tree that mimics a dual-socket server.
func buildFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	fixture(t, root, "/proc/cpuinfo", cpuInfoTwoSockets)
	fixture(t, root, "/proc/meminfo", "MemTotal:       1030804372 kB\nMemFree:         123456 kB\n")
	fixture(t, root, "/proc/uptime", "482913.44 1902333.01\n")
	fixture(t, root, "/proc/cmdline", "BOOT_IMAGE=/vmlinuz root=/dev/mapper/root ro quiet\n")
	fixture(t, root, "/proc/sys/kernel/numa_balancing", "1\n")
	fixture(t, root, "/etc/hostname", "node-014\n")

	fixture(t, root, "/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor", "performance\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cpufreq/base_frequency", "2000000\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq", "3500000\n")

	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index0/level", "1\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index0/type", "Data\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index0/size", "48K\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index0/shared_cpu_list", "0-27\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index1/level", "1\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index1/type", "Instruction\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index1/size", "32K\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index1/shared_cpu_list", "0-27\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index2/level", "2\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index2/type", "Unified\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index2/size", "2048K\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index2/shared_cpu_list", "0-27\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index3/level", "3\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index3/type", "Unified\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index3/size", "107520K\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/cache/index3/shared_cpu_list", "0-55\n")

	fixture(t, root, "/sys/devices/system/cpu/cpu0/thermal_throttle/core_throttle_count", "0\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/thermal_throttle/package_throttle_count", "0\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/thermal_throttle/package_power_limit_count", "17\n")

	fixture(t, root, "/sys/devices/system/node/node0/cpulist", "0-27\n")
	fixture(t, root, "/sys/devices/system/node/node0/meminfo", "Node 0 MemTotal:       515402186 kB\n")
	fixture(t, root, "/sys/devices/system/node/node1/cpulist", "28-55\n")
	fixture(t, root, "/sys/devices/system/node/node1/meminfo", "Node 1 MemTotal:       515402186 kB\n")

	fixture(t, root, "/sys/kernel/mm/transparent_hugepage/enabled", "always [madvise] never\n")
	fixture(t, root, "/sys/module/pcie_aspm/parameters/policy", "[default] performance powersave powersupersave\n")
	fixture(t, root, "/sys/devices/system/cpu/cpu0/power/energy_perf_bias", "performance\n")
	fixture(t, root, "/sys/devices/system/cpu/vulnerabilities/spectre_v2", "Mitigation: Enhanced IBRS\n")
	fixture(t, root, "/sys/devices/system/cpu/vulnerabilities/spec_rstack_overflow", "Vulnerable: Safe RET, no microcode\n")

	// The firmware table.
	b := &dmitest.Builder{}
	b.BIOS("Dell Inc.", "2.10.1", "03/14/2024")
	b.System("Dell Inc.", "PowerEdge R760", "Not Specified", "CN7792", "4c4c4544", "0KJ3T2")
	b.Baseboard("Dell Inc.", "0KJ3T2", "A00")
	b.MemoryArray(0x1100, 1<<30, 16)
	for i, ch := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
		for slot := 1; slot <= 2; slot++ {
			b.MemoryDevice(dmitest.MemorySpec{
				Handle:        uint16(0x1200 + i*2 + slot),
				ArrayHandle:   0x1100,
				Locator:       "DIMM_" + ch + string(rune('0'+slot)),
				BankLocator:   "NODE 0 " + ch,
				SizeMB:        64 * 1024,
				MemoryType:    0x1B,
				Vendor:        "Kingston",
				Serial:        "0A1B2C3D",
				PartNumber:    "KSM56R46BD4-256HA",
				RankCount:     2,
				SpeedMTs:      5600,
				ConfiguredMTs: 4800,
			})
		}
	}
	fixture(t, root, "/sys/firmware/dmi/tables/DMI", string(b.Build()))

	return root
}

// writePCI writes the PCI bus fixture.
//
// PCI addresses contain colons, which are not valid in Windows path names,
// so this part of the fixture only exists on the platform the tool targets.
func writePCI(t *testing.T, root string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("PCI address directory names are not valid on Windows")
	}

	// A GPU that negotiated a degraded link, and the upstream port it sits on.
	fixture(t, root, "/sys/bus/pci/devices/0000:00:1c.0/vendor", "0x8086\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:00:1c.0/device", "0xa1d8\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:00:1c.0/class", "0x060400\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:00:1c.0/max_link_speed", "32.0 GT/s PCIe\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:00:1c.0/max_link_width", "16\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:00:1c.0/current_link_speed", "32.0 GT/s PCIe\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:00:1c.0/current_link_width", "16\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:00:1c.0/numa_node", "0\n")

	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/vendor", "0x10de\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/device", "0x2684\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/class", "0x030200\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/numa_node", "0\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/max_link_speed", "32.0 GT/s PCIe\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/max_link_width", "16\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/current_link_speed", "16.0 GT/s PCIe\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/current_link_width", "16\n")
	fixture(t, root, "/sys/bus/pci/devices/0000:01:00.0/aer_dev_correctable", "42\n")

	fixture(t, root, "/sys/devices/pci0000:00/0000:00:1c.0/0000:01:00.0/.keep", "")
}

func TestGatherReadsTheMachine(t *testing.T) {
	root := buildFixture(t)
	m := Gather(sysfs.New(root))

	if m.CPU.Sockets != 2 {
		t.Errorf("sockets = %d, want 2", m.CPU.Sockets)
	}
	if m.CPU.CoresPerSocket != 28 {
		t.Errorf("cores per socket = %d, want 28", m.CPU.CoresPerSocket)
	}
	if m.CPU.ThreadsPerCore != 2 {
		t.Errorf("threads per core = %d, want 2", m.CPU.ThreadsPerCore)
	}
	if !m.CPU.SMTEnabled {
		t.Error("SMT should be enabled")
	}
	if m.CPU.ISA.Count == 0 || m.CPU.ISA.Fingerprint == "" {
		t.Error("ISA feature set was not collected")
	}
	if m.CPU.BaseFreqMHz != 2000 || m.CPU.MaxFreqMHz != 3500 {
		t.Errorf("frequencies = %d/%d MHz, want 2000/3500", m.CPU.BaseFreqMHz, m.CPU.MaxFreqMHz)
	}
	if !m.CPU.TurboEnabled {
		t.Error("turbo should be reported as available")
	}
	if m.CPU.Cache.L3KB != 107520 {
		t.Errorf("L3 = %d KB, want 107520", m.CPU.Cache.L3KB)
	}
}

func TestGatherReadsMemoryPopulation(t *testing.T) {
	root := buildFixture(t)
	m := Gather(sysfs.New(root))

	if m.Memory.Source != "smbios" {
		t.Fatalf("memory source = %q, want smbios", m.Memory.Source)
	}
	if len(m.Memory.Channels) != 8 {
		t.Fatalf("channels = %d, want 8", len(m.Memory.Channels))
	}
	if m.Memory.SlotsFilled != 16 {
		t.Errorf("slots filled = %d, want 16", m.Memory.SlotsFilled)
	}
	if m.Memory.SlotsTotal != 16 {
		t.Errorf("slots total = %d, want 16", m.Memory.SlotsTotal)
	}
	if m.Memory.ConfiguredMTs != 4800 {
		t.Errorf("configured speed = %d MT/s, want 4800", m.Memory.ConfiguredMTs)
	}
	if m.Memory.MaxMTs != 5600 {
		t.Errorf("rated speed = %d MT/s, want 5600", m.Memory.MaxMTs)
	}
	// Eight channels at 4800 MT/s move 8 bytes per transfer each.
	if want := 307.2; m.Memory.BandwidthGBs < want-0.01 || m.Memory.BandwidthGBs > want+0.01 {
		t.Errorf("bandwidth = %.1f GB/s, want %.1f", m.Memory.BandwidthGBs, want)
	}
	if want := 358.4; m.Memory.BandwidthAtMaxGBs < want-0.01 {
		t.Errorf("bandwidth at rated speed = %.1f GB/s, want %.1f", m.Memory.BandwidthAtMaxGBs, want)
	}
}

func TestGatherReadsPCIeLinkState(t *testing.T) {
	root := buildFixture(t)
	writePCI(t, root)
	m := Gather(sysfs.New(root))

	if len(m.PCIe.Devices) != 2 {
		t.Fatalf("pcie devices = %d, want 2", len(m.PCIe.Devices))
	}
	var gpu *PCIeDevice
	for i := range m.PCIe.Devices {
		if m.PCIe.Devices[i].Address == "0000:01:00.0" {
			gpu = &m.PCIe.Devices[i]
		}
	}
	if gpu == nil {
		t.Fatal("the endpoint was not collected")
	}
	if !gpu.Degraded {
		t.Error("the endpoint should be reported as degraded")
	}
	if gpu.CurrentSpeedGTs != 16 || gpu.MaxSpeedGTs != 32 {
		t.Errorf("link speeds = %.1f of %.1f GT/s, want 16 of 32", gpu.CurrentSpeedGTs, gpu.MaxSpeedGTs)
	}
	if gpu.PortMaxSpeedGTs != 32 {
		t.Errorf("upstream port limit = %.1f GT/s, want 32", gpu.PortMaxSpeedGTs)
	}
	if gpu.AERCorrectable != 42 {
		t.Errorf("correctable errors = %d, want 42", gpu.AERCorrectable)
	}
	if m.PCIe.Degraded != 1 {
		t.Errorf("degraded count = %d, want 1", m.PCIe.Degraded)
	}
}

func TestGatherDegradesWithoutSources(t *testing.T) {
	root := t.TempDir()
	fixture(t, root, "/proc/cpuinfo", cpuInfoTwoSockets)
	m := Gather(sysfs.New(root))

	if m.Memory.Source != "sysfs" {
		t.Errorf("memory source = %q, want sysfs", m.Memory.Source)
	}
	if len(m.Notes) == 0 {
		t.Error("missing sources should be reported as notes")
	}
	if m.CPU.Sockets != 2 {
		t.Errorf("sockets = %d, want 2", m.CPU.Sockets)
	}
}

func TestGatherReadsNUMAAndThrottling(t *testing.T) {
	root := buildFixture(t)
	m := Gather(sysfs.New(root))

	if m.NUMA.Nodes != 2 {
		t.Errorf("numa nodes = %d, want 2", m.NUMA.Nodes)
	}
	if m.NUMA.Interleaved {
		t.Error("two nodes must not be reported as interleaved")
	}
	if m.Power.ThrottlePower != 17 {
		t.Errorf("power throttle events = %d, want 17", m.Power.ThrottlePower)
	}
	if m.OS.NUMABalancing != "1" {
		t.Errorf("numa balancing = %q, want 1", m.OS.NUMABalancing)
	}
}
