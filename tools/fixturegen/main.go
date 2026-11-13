// Command fixturegen writes a machine fixture to disk.
//
// The fixture is a synthetic but faithful copy of what a dual-socket server
// exposes: a CPU with two sockets and SMT, a firmware table describing eight
// DDR5 channels running below the rating of the modules, a PCIe link that
// negotiated a generation too low, and a power limit throttle that already
// happened.
//
// It lets anyone try asbuilt without a server and without root:
//
//	go run ./tools/fixturegen -out tests/fixtures/server
//	asbuilt --root tests/fixtures/server check
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Balanov/asbuilt/internal/dmitest"
)

func main() {
	out := flag.String("out", "tests/fixtures/server", "directory to write the fixture to")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "fixturegen:", err)
		os.Exit(1)
	}

	files := map[string]string{
		"/etc/hostname": "node-014\n",
		"/proc/cmdline": "BOOT_IMAGE=/vmlinuz-6.1.0-42-cloud-amd64 root=/dev/mapper/root ro quiet mitigations=auto\n",
		"/proc/meminfo": "MemTotal:       1030804372 kB\nMemFree:         185441240 kB\n",
		"/proc/uptime":  "482913.44 1902333.01\n",
		"/proc/version": "Linux version 6.1.0-42-cloud-amd64 (debian-kernel@lists.debian.org) (gcc-12 (Debian 12.2.0-14) 12.2.0, GNU ld 2.40) #1 SMP PREEMPT_DYNAMIC Debian 6.1.102-3 (2024-11-22)\n",
		"/proc/sys/kernel/numa_balancing": "1\n",

		"/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor": "performance\n",
		"/sys/devices/system/cpu/cpu0/cpufreq/base_frequency":  "2000000\n",
		"/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq": "3500000\n",
		"/sys/devices/system/cpu/cpu0/power/energy_perf_bias":   "performance\n",

		"/sys/devices/system/cpu/cpu0/cache/index0/level":            "1\n",
		"/sys/devices/system/cpu/cpu0/cache/index0/type":             "Data\n",
		"/sys/devices/system/cpu/cpu0/cache/index0/size":             "48K\n",
		"/sys/devices/system/cpu/cpu0/cache/index0/shared_cpu_list":  "0-27\n",
		"/sys/devices/system/cpu/cpu0/cache/index1/level":            "1\n",
		"/sys/devices/system/cpu/cpu0/cache/index1/type":             "Instruction\n",
		"/sys/devices/system/cpu/cpu0/cache/index1/size":             "32K\n",
		"/sys/devices/system/cpu/cpu0/cache/index1/shared_cpu_list":  "0-27\n",
		"/sys/devices/system/cpu/cpu0/cache/index2/level":            "2\n",
		"/sys/devices/system/cpu/cpu0/cache/index2/type":             "Unified\n",
		"/sys/devices/system/cpu/cpu0/cache/index2/size":             "2048K\n",
		"/sys/devices/system/cpu/cpu0/cache/index2/shared_cpu_list":  "0-27\n",
		"/sys/devices/system/cpu/cpu0/cache/index3/level":            "3\n",
		"/sys/devices/system/cpu/cpu0/cache/index3/type":             "Unified\n",
		"/sys/devices/system/cpu/cpu0/cache/index3/size":             "107520K\n",
		"/sys/devices/system/cpu/cpu0/cache/index3/shared_cpu_list":  "0-55\n",

		"/sys/devices/system/cpu/cpu0/thermal_throttle/core_throttle_count":          "0\n",
		"/sys/devices/system/cpu/cpu0/thermal_throttle/package_throttle_count":       "0\n",
		"/sys/devices/system/cpu/cpu0/thermal_throttle/package_power_limit_count":   "17\n",
		"/sys/kernel/mm/transparent_hugepage/enabled":                                "always [madvise] never\n",
		"/sys/module/pcie_aspm/parameters/policy":                                    "[default] performance powersave powersupersave\n",
		"/sys/devices/system/cpu/vulnerabilities/spectre_v2":                          "Mitigation: Enhanced IBRS\n",
		"/sys/devices/system/cpu/vulnerabilities/spec_rstack_overflow":                "Vulnerable: Safe RET, no microcode\n",

		"/sys/devices/system/node/node0/cpulist": "0-27\n",
		"/sys/devices/system/node/node0/meminfo": "Node 0 MemTotal:       515402186 kB\n",
		"/sys/devices/system/node/node1/cpulist": "28-55\n",
		"/sys/devices/system/node/node1/meminfo": "Node 1 MemTotal:       515402186 kB\n",
	}

	// The processor list: two sockets, twenty-eight cores each, two threads.
	var cpuinfo strings.Builder
	for socket := 0; socket < 2; socket++ {
		for proc := 0; proc < 56; proc++ {
			fmt.Fprintf(&cpuinfo, "processor\t: %d\n", socket*56+proc)
			cpuinfo.WriteString("vendor_id\t: GenuineIntel\n")
			cpuinfo.WriteString("cpu family\t: 6\n")
			cpuinfo.WriteString("model\t\t: 143\n")
			cpuinfo.WriteString("model name\t: Intel(R) Xeon(R) Platinum 8480+\n")
			cpuinfo.WriteString("stepping\t: 8\n")
			cpuinfo.WriteString("microcode\t: 0x2b000603\n")
			cpuinfo.WriteString("cpu MHz\t\t: 2000.000\n")
			cpuinfo.WriteString("cache size\t: 107520 KB\n")
			fmt.Fprintf(&cpuinfo, "physical id\t: %d\n", socket)
			cpuinfo.WriteString("siblings\t: 56\n")
			fmt.Fprintf(&cpuinfo, "core id\t\t: %d\n", proc%28)
			cpuinfo.WriteString("cpu cores\t: 28\n")
			cpuinfo.WriteString("flags\t\t: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology cpuid tsc_known_freq pni pclmulqdq ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt aes xsave avx f16c rdrand hypervisor lahf_lm abm 3dnowprefetch invpcid_single ssbd ibrs ibpb stibp fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid mpx rdseed adx smap clflushopt xsaveopt xsavec xgetbv1 arat\n")
			cpuinfo.WriteString("bugs\t\t:\n")
			cpuinfo.WriteString("bogomips\t: 4000.00\n")
			cpuinfo.WriteString("TLB size\t: 0\n")
			cpuinfo.WriteString("address sizes\t: 46 bits physical, 57 bits virtual\n")
			cpuinfo.WriteString("power management:\n\n")
		}
	}
	files["/proc/cpuinfo"] = cpuinfo.String()

	// The firmware table: a Dell PowerEdge R760 with eight DDR5 channels.
	b := &dmitest.Builder{}
	b.BIOS("Dell Inc.", "2.10.1", "03/14/2024")
	b.System("Dell Inc.", "PowerEdge R760", "Not Specified", "CN7792", "4c4c4544-4343", "0KJ3T2")
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
	files["/sys/firmware/dmi/tables/DMI"] = string(b.Build())

	// The PCI bus: an upstream port that allows Gen5 x16 and an accelerator
	// that negotiated Gen4 x16.
	files["/sys/bus/pci/devices/0000:00:1c.0/vendor"] = "0x8086\n"
	files["/sys/bus/pci/devices/0000:00:1c.0/device"] = "0xa1d8\n"
	files["/sys/bus/pci/devices/0000:00:1c.0/class"] = "0x060400\n"
	files["/sys/bus/pci/devices/0000:00:1c.0/max_link_speed"] = "32.0 GT/s PCIe\n"
	files["/sys/bus/pci/devices/0000:00:1c.0/max_link_width"] = "16\n"
	files["/sys/bus/pci/devices/0000:00:1c.0/current_link_speed"] = "32.0 GT/s PCIe\n"
	files["/sys/bus/pci/devices/0000:00:1c.0/current_link_width"] = "16\n"
	files["/sys/bus/pci/devices/0000:00:1c.0/numa_node"] = "0\n"

	files["/sys/bus/pci/devices/0000:01:00.0/vendor"] = "0x10de\n"
	files["/sys/bus/pci/devices/0000:01:00.0/device"] = "0x2684\n"
	files["/sys/bus/pci/devices/0000:01:00.0/class"] = "0x030200\n"
	files["/sys/bus/pci/devices/0000:01:00.0/numa_node"] = "0\n"
	files["/sys/bus/pci/devices/0000:01:00.0/max_link_speed"] = "32.0 GT/s PCIe\n"
	files["/sys/bus/pci/devices/0000:01:00.0/max_link_width"] = "16\n"
	files["/sys/bus/pci/devices/0000:01:00.0/current_link_speed"] = "16.0 GT/s PCIe\n"
	files["/sys/bus/pci/devices/0000:01:00.0/current_link_width"] = "16\n"
	files["/sys/bus/pci/devices/0000:01:00.0/aer_dev_correctable"] = "42\n"
	files["/sys/bus/pci/devices/0000:01:00.0/aer_dev_fatal"] = "0\n"

	files["/sys/devices/pci0000:00/0000:00:1c.0/0000:01:00.0/.keep"] = ""

	for path, content := range files {
		full := filepath.Join(*out, filepath.FromSlash(strings.TrimPrefix(path, "/")))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "fixturegen:", err)
			os.Exit(1)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "fixturegen:", err)
			os.Exit(1)
		}
	}
	fmt.Printf("wrote %d files to %s\n", len(files), *out)
}
