package rules

import (
	"fmt"
	"strings"
)

// cpuTurboDisabled reports a processor whose turbo range is closed.
func cpuTurboDisabled() Rule { return Rule{
	ID:       "cpu.turbo.disabled",
	Title:    "CPU turbo is not available",
	Severity: Warn,
	Source:   "https://www.intel.com/content/www/us/en/support/articles/000057277.html",
	Remedy:   "Enable turbo in firmware and check the power limits: PL1 equal to PL2 removes the headroom.",
	Check: func(c *Context) []Finding {
		cpu := c.Manifest.CPU
		if cpu.BaseFreqMHz <= 0 || cpu.MaxFreqMHz <= 0 || cpu.MaxFreqMHz > cpu.BaseFreqMHz {
			return nil
		}
		return []Finding{finding(cpuTurboDisabled(),
			fmt.Sprintf("the frequency policy tops out at the base clock of %d MHz", cpu.BaseFreqMHz),
			fmt.Sprintf("base_frequency=%d MHz", cpu.BaseFreqMHz),
			fmt.Sprintf("scaling_max_freq=%d MHz", cpu.MaxFreqMHz))}
	},
}
}

// cpuGovernorNotPerformance reports a governor that holds the clock down.
func cpuGovernorNotPerformance() Rule { return Rule{
	ID:       "cpu.governor.not_performance",
	Title:    "CPU frequency governor is not a performance governor",
	Severity: Warn,
	Source:   "https://documentation.suse.com/sles/15-SP7/single-html/SLES-tuning/index.html",
	Remedy:   "Set the governor to performance for latency-bound workloads.",
	Check: func(c *Context) []Finding {
		gov := c.Manifest.CPU.Governor
		switch gov {
		case "powersave", "conservative", "schedutil", "ondemand":
		default:
			return nil
		}
		return []Finding{finding(cpuGovernorNotPerformance(),
			fmt.Sprintf("the governor is %q, so the clock is scaled rather than held", gov),
			fmt.Sprintf("scaling_governor=%s", gov))}
	},
}
}

// cpuEnergyPerfBiasNotPerformance reports an energy bias that trades
// performance for power.
func cpuEnergyPerfBiasNotPerformance() Rule { return Rule{
	ID:       "cpu.energy_perf_bias.not_performance",
	Title:    "CPU energy performance bias is not set to performance",
	Severity: Info,
	Source:   "https://www.kernel.org/doc/html/latest/admin-guide/pm/amd-pstate.html",
	Remedy:   "Set the bias to performance when the workload is latency-bound.",
	Check: func(c *Context) []Finding {
		bias := strings.TrimSpace(c.Manifest.OS.EnergyPerfBias)
		if bias == "" || bias == "performance" {
			return nil
		}
		return []Finding{finding(cpuEnergyPerfBiasNotPerformance(),
			fmt.Sprintf("the energy performance bias is %q", bias),
			fmt.Sprintf("energy_perf_bias=%s", bias))}
	},
}
}

// cpuSMTDisabled reports that hardware threads are switched off.
func cpuSMTDisabled() Rule { return Rule{
	ID:       "cpu.smt.disabled",
	Title:    "Simultaneous multithreading is disabled",
	Severity: Info,
	Source:   "https://www.intel.com/content/www/us/en/developer/articles/technical/measuring-impact-of-numa-migrations-on-performance.html",
	Remedy:   "Confirm this is intended: SMT off changes both licensing and throughput characteristics.",
	Check: func(c *Context) []Finding {
		cpu := c.Manifest.CPU
		if cpu.LogicalCPUs <= 0 {
			return nil
		}
		if cpu.ThreadsPerCore <= 1 {
			return []Finding{finding(cpuSMTDisabled(),
				fmt.Sprintf("the %d logical CPUs are %d cores without hardware threads", cpu.LogicalCPUs, cpu.CoresPerSocket*cpu.Sockets),
				fmt.Sprintf("threads_per_core=%d", cpu.ThreadsPerCore))}
		}
		return nil
	},
}
}

// osTHPAlways reports transparent huge pages switched on for every
// allocation, which costs latency.
func osTHPAlways() Rule { return Rule{
	ID:       "os.thp.always",
	Title:    "Transparent huge pages are enabled for all allocations",
	Severity: Warn,
	Source:   "https://documentation.suse.com/sles/15-SP7/single-html/SLES-tuning/index.html",
	Remedy:   "Use the madvise mode and let the workload request huge pages explicitly.",
	Check: func(c *Context) []Finding {
		thp := c.Manifest.OS.THP
		if thp == "" || !strings.Contains(thp, "always") {
			return nil
		}
		mode := thp
		if i := strings.Index(thp, "["); i >= 0 {
			mode = thp[i+1:]
			if j := strings.Index(mode, "]"); j >= 0 {
				mode = mode[:j]
			}
		}
		if mode != "always" {
			return nil
		}
		return []Finding{finding(osTHPAlways(),
			"transparent huge pages are always on, so every eligible mapping is backed by a huge page",
			fmt.Sprintf("transparent_hugepage=%s", thp))}
	},
}
}

// osNUMABalancingEnabled reports automatic NUMA balancing.
func osNUMABalancingEnabled() Rule { return Rule{
	ID:       "os.numa_balancing.enabled",
	Title:    "Automatic NUMA balancing is active",
	Severity: Info,
	Source:   "https://documentation.suse.com/sles/15-SP7/single-html/SLES-tuning/index.html",
	Remedy:   "Leave it on unless the workload pins its memory deliberately.",
	Check: func(c *Context) []Finding {
		if strings.TrimSpace(c.Manifest.OS.NUMABalancing) != "1" {
			return nil
		}
		return []Finding{finding(osNUMABalancingEnabled(),
			"the kernel migrates pages between nodes on its own, which adds a fault to the critical path",
			"numa_balancing=1")}
	},
}
}

// osASPMPolicyNotPerformance reports an aggressive link power policy, the
// usual suspect when a link keeps training down.
func osASPMPolicyNotPerformance() Rule { return Rule{
	ID:       "os.pcie_aspm.not_performance",
	Title:    "PCIe link power management is not set to performance",
	Severity: Info,
	Source:   "https://onestopsystems.com/blogs/one-stop-systems-blog/beyond-lspci-how-to-debug-maintain-high-density-pcie-gen5-gen6-fabrics-at-the-edge",
	Remedy:   "Set the ASPM policy to performance while diagnosing a link that trains down.",
	Check: func(c *Context) []Finding {
		policy := selectedMode(c.Manifest.OS.ASPM)
		if policy == "" || policy == "performance" {
			return nil
		}
		return []Finding{finding(osASPMPolicyNotPerformance(),
			fmt.Sprintf("the ASPM policy is %q, so links may drop to a lower speed when idle", policy),
			fmt.Sprintf("pcie_aspm policy=%s", policy))}
	},
}
}

// osMitigationVulnerable reports a mitigation the kernel itself marks as
// not in place.
func osMitigationVulnerable() Rule { return Rule{
	ID:       "os.mitigation.vulnerable",
	Title:    "The kernel reports a vulnerability without mitigation",
	Severity: Warn,
	Source:   "https://www.kernel.org/doc/html/latest/admin-guide/hw-vuln/index.html",
	Remedy:   "Update the kernel or enable the mitigation the vendor documents for this CPU.",
	Check: func(c *Context) []Finding {
		var out []Finding
		names := make([]string, 0, len(c.Manifest.OS.Mitigations))
		for name := range c.Manifest.OS.Mitigations {
			names = append(names, name)
		}
		sortStrings(names)
		for _, name := range names {
			status := c.Manifest.OS.Mitigations[name]
			if !strings.HasPrefix(status, "Vulnerable") {
				continue
			}
			out = append(out, finding(osMitigationVulnerable(),
				fmt.Sprintf("%s: %s", name, status),
				fmt.Sprintf("%s=%s", name, status)))
		}
		return out
	},
}
}

// powerThermalThrottled reports thermal throttling that already happened.
func powerThermalThrottled() Rule { return Rule{
	ID:       "power.throttle.thermal",
	Title:    "The processor has been thermally throttled",
	Severity: Critical,
	Source:   "https://www.suse.com/support/update/announcement/2024/suse-su-20240224-1/",
	Remedy:   "Check the cooling path, the ambient temperature and the thermal limits in firmware.",
	Check: func(c *Context) []Finding {
		n := c.Manifest.Power.ThrottleThermal
		if n == 0 {
			return nil
		}
		return []Finding{finding(powerThermalThrottled(),
			fmt.Sprintf("%d thermal throttle events were recorded since boot", n),
			fmt.Sprintf("thermal_throttle counters=%d", n))}
	},
}
}

// powerLimitThrottled reports power limit throttling that already happened.
func powerLimitThrottled() Rule { return Rule{
	ID:       "power.throttle.limit",
	Title:    "The processor has been power limit throttled",
	Severity: Critical,
	Source:   "https://www.intel.com/content/www/us/en/support/articles/000057277.html",
	Remedy:   "Raise PL1 in firmware or reduce the workload: the package is spending its budget faster than the budget allows.",
	Check: func(c *Context) []Finding {
		n := c.Manifest.Power.ThrottlePower
		if n == 0 {
			return nil
		}
		return []Finding{finding(powerLimitThrottled(),
			fmt.Sprintf("%d power limit throttle events were recorded since boot", n),
			fmt.Sprintf("package_power_limit_count=%d", n))}
	},
}
}

func sortStrings(v []string) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j] < v[j-1]; j-- {
			v[j], v[j-1] = v[j-1], v[j]
		}
	}
}

// selectedMode extracts the active value from a kernel selector file such as
// "always [madvise] never" or "[default] performance powersave powersupersave".
func selectedMode(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.Index(s, "["); i >= 0 {
		rest := s[i+1:]
		if j := strings.Index(rest, "]"); j >= 0 {
			return strings.TrimSpace(rest[:j])
		}
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
