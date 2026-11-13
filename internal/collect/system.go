package collect

import (
	"strconv"
	"strings"

	"github.com/Balanov/asbuilt/internal/sysfs"
)

// collectSystem fills the OS, NUMA, Power and virtualization sections.
func collectSystem(f *sysfs.FS, m *Manifest) {
	if v, err := f.ReadString("/proc/version"); err == nil {
		m.OS.Kernel = firstLine(v)
	}
	if cmdline, err := f.ReadString("/proc/cmdline"); err == nil && cmdline != "" {
		m.OS.KernelCmdline = strings.Fields(cmdline)
	}
	if thp, err := f.ReadString("/sys/kernel/mm/transparent_hugepage/enabled"); err == nil {
		m.OS.THP = thp
	}
	if bal, err := f.ReadString("/proc/sys/kernel/numa_balancing"); err == nil {
		m.OS.NUMABalancing = bal
		m.NUMA.AutoBalancing = bal
	}
	if aspm, err := f.ReadString("/sys/module/pcie_aspm/parameters/policy"); err == nil {
		m.OS.ASPM = aspm
	}
	if bias, err := f.ReadString("/sys/devices/system/cpu/cpu0/power/energy_perf_bias"); err == nil {
		m.OS.EnergyPerfBias = bias
	}
	m.OS.Mitigations = readVulnerabilities(f)

	collectNUMA(f, m)
	collectPower(f, m)
	collectVirt(f, m)
}

// readVulnerabilities reads the kernel's mitigation status table.
func readVulnerabilities(f *sysfs.FS) map[string]string {
	names := f.List("/sys/devices/system/cpu/vulnerabilities")
	if len(names) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, name := range names {
		if v, err := f.ReadString("/sys/devices/system/cpu/vulnerabilities/" + name); err == nil {
			out[name] = v
		}
	}
	return out
}

// collectNUMA reads the memory topology from sysfs.
func collectNUMA(f *sysfs.FS, m *Manifest) {
	nodes := f.List("/sys/devices/system/node")
	for _, name := range nodes {
		if !strings.HasPrefix(name, "node") {
			continue
		}
		id, err := strconv.Atoi(strings.TrimPrefix(name, "node"))
		if err != nil {
			continue
		}
		n := NUMANode{ID: id}
		if cpus, err := f.ReadString("/sys/devices/system/node/" + name + "/cpulist"); err == nil {
			n.CPUs = cpus
		}
		if info, err := f.ReadString("/sys/devices/system/node/" + name + "/meminfo"); err == nil {
			n.MemTotalKB = memTotalKB(info)
		}
		m.NUMA.Topology = append(m.NUMA.Topology, n)
	}
	m.NUMA.Nodes = len(m.NUMA.Topology)
	// A multi-socket machine that presents a single memory node is running
	// with node interleaving, which hides the real topology from the OS.
	m.NUMA.Interleaved = m.NUMA.Nodes <= 1 && m.CPU.Sockets > 1
}

// collectPower reads the throttling counters and the RAPL energy counters.
func collectPower(f *sysfs.FS, m *Manifest) {
	base := "/sys/devices/system/cpu/cpu0/thermal_throttle"
	m.Power.ThrottleThermal = parseUint(readString(f, base+"/core_throttle_count")) +
		parseUint(readString(f, base+"/package_throttle_count"))
	m.Power.ThrottlePower = parseUint(readString(f, base+"/package_power_limit_count"))

	for _, name := range f.List("/sys/class/powercap") {
		if !strings.HasPrefix(name, "intel-rapl") && !strings.HasPrefix(name, "amd-rapl") {
			continue
		}
		dir := "/sys/class/powercap/" + name
		if n, err := f.ReadString(dir + "/name"); err == nil && strings.Contains(n, "package") {
			m.Power.RaplPackages++
			m.Power.RaplName = n
		}
		if uj, err := f.ReadString(dir + "/energy_uj"); err == nil {
			m.Power.EnergyJoules += float64(atoiSafe(uj)) / 1e6
		}
	}
}

// collectVirt detects the virtualization layer.
func collectVirt(f *sysfs.FS, m *Manifest) {
	if raw, err := f.ReadFile("/proc/cpuinfo"); err == nil {
		if strings.Contains(string(raw), "hypervisor") {
			m.Virt.Hypervisor = "present"
		}
	}
	if t, err := f.ReadString("/sys/hypervisor/type"); err == nil && t != "" {
		m.Virt.Type = t
		m.Virt.Hypervisor = t
	}
	switch {
	case strings.Contains(strings.ToLower(m.Plat.Vendor), "vmware"):
		m.Virt.Hypervisor = "VMware"
	case strings.Contains(strings.ToLower(m.Plat.Vendor), "microsoft"):
		m.Virt.Hypervisor = "Microsoft"
	case strings.Contains(strings.ToLower(m.Plat.Vendor), "xen"):
		m.Virt.Hypervisor = "Xen"
	case strings.Contains(strings.ToLower(m.Plat.Vendor), "amazon"):
		m.Virt.Hypervisor = "Amazon EC2"
	case strings.Contains(strings.ToLower(m.Plat.Vendor), "google"):
		m.Virt.Hypervisor = "Google Compute Engine"
	case strings.Contains(strings.ToLower(m.Plat.Vendor), "qemu"):
		m.Virt.Hypervisor = "QEMU/KVM"
	case strings.Contains(strings.ToLower(m.Plat.Vendor), "bochs"):
		m.Virt.Hypervisor = "QEMU"
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}
