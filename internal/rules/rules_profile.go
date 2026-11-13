package rules

import (
	"fmt"
	"strings"

	"github.com/Balanov/asbuilt/internal/collect"
	"github.com/Balanov/asbuilt/internal/profile"
)

// Profile is the expected configuration a machine is checked against.
type Profile = profile.Profile

// profilePlatformMismatch reports a machine that is not the machine the
// profile describes.
func profilePlatformMismatch() Rule { return Rule{
	ID:       "profile.platform.mismatch",
	Title:    "The platform does not match the expected profile",
	Severity: Critical,
	Source:   "https://github.com/intel/PerfSpect",
	Remedy:   "Check that the node was built and provisioned as the fleet intended.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.Platform == nil {
			return nil
		}
		e := p.Expect.Platform
		var out []Finding
		m := c.Manifest
		if e.Vendor != "" && !strings.EqualFold(e.Vendor, m.Plat.Vendor) {
			out = append(out, finding(profilePlatformMismatch(),
				fmt.Sprintf("vendor is %q, expected %q", m.Plat.Vendor, e.Vendor),
				fmt.Sprintf("system manufacturer=%s", m.Plat.Vendor)))
		}
		if e.Product != "" && !strings.EqualFold(e.Product, m.Plat.Product) {
			out = append(out, finding(profilePlatformMismatch(),
				fmt.Sprintf("product is %q, expected %q", m.Plat.Product, e.Product),
				fmt.Sprintf("system product=%s", m.Plat.Product)))
		}
		if e.BIOSVersion != "" && !strings.EqualFold(e.BIOSVersion, m.Plat.BIOS.Version) {
			out = append(out, finding(profilePlatformMismatch(),
				fmt.Sprintf("BIOS version is %q, expected %q", m.Plat.BIOS.Version, e.BIOSVersion),
				fmt.Sprintf("bios version=%s", m.Plat.BIOS.Version)))
		}
		return out
	},
}
}

// profileCPUSockets reports a different number of sockets than expected.
func profileCPUSockets() Rule { return Rule{
	ID:       "profile.cpu.sockets",
	Title:    "Socket count does not match the expected profile",
	Severity: Critical,
	Source:   "https://cray-hpe.github.io/docs-sat/en-25/introduction/",
	Remedy:   "Replace the node or correct the profile.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.CPU == nil || p.Expect.CPU.Sockets == nil {
			return nil
		}
		got := c.Manifest.CPU.Sockets
		want := *p.Expect.CPU.Sockets
		if got == want {
			return nil
		}
		return []Finding{finding(profileCPUSockets(),
			fmt.Sprintf("the machine has %d sockets, expected %d", got, want),
			fmt.Sprintf("sockets=%d", got))}
	},
}
}

// profileCPUThreads reports a different SMT layout than expected.
func profileCPUThreads() Rule { return Rule{
	ID:       "profile.cpu.threads_per_core",
	Title:    "Threads per core do not match the expected profile",
	Severity: Warn,
	Source:   "https://cray-hpe.github.io/docs-sat/en-25/introduction/",
	Remedy:   "Check the SMT setting in firmware.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.CPU == nil || p.Expect.CPU.ThreadsPerCore == nil {
			return nil
		}
		got := c.Manifest.CPU.ThreadsPerCore
		want := *p.Expect.CPU.ThreadsPerCore
		if got == want {
			return nil
		}
		return []Finding{finding(profileCPUThreads(),
			fmt.Sprintf("the machine runs %d threads per core, expected %d", got, want),
			fmt.Sprintf("threads_per_core=%d", got))}
	},
}
}

// profileCPUTurbo reports a turbo state that differs from the profile.
func profileCPUTurbo() Rule { return Rule{
	ID:       "profile.cpu.turbo",
	Title:    "Turbo state does not match the expected profile",
	Severity: Warn,
	Source:   "https://cray-hpe.github.io/docs-sat/en-25/introduction/",
	Remedy:   "Check the turbo setting and the power limits in firmware.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.CPU == nil || p.Expect.CPU.TurboEnabled == nil {
			return nil
		}
		got := c.Manifest.CPU.TurboEnabled
		want := *p.Expect.CPU.TurboEnabled
		if got == want {
			return nil
		}
		state := "disabled"
		if got {
			state = "enabled"
		}
		return []Finding{finding(profileCPUTurbo(),
			fmt.Sprintf("turbo is %s, expected the opposite", state),
			fmt.Sprintf("turbo_enabled=%t", got))}
	},
}
}

// profileMemoryChannels reports a different number of populated channels.
func profileMemoryChannels() Rule { return Rule{
	ID:       "profile.memory.channels",
	Title:    "Populated memory channels do not match the expected profile",
	Severity: Critical,
	Source:   "https://cray-hpe.github.io/docs-sat/en-25/introduction/",
	Remedy:   "Repopulate the node or correct the profile.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.Memory == nil || p.Expect.Memory.Channels == nil {
			return nil
		}
		got := len(c.Manifest.Memory.Channels)
		want := *p.Expect.Memory.Channels
		if got == want {
			return nil
		}
		return []Finding{finding(profileMemoryChannels(),
			fmt.Sprintf("%d memory channels are populated, expected %d", got, want),
			fmt.Sprintf("channels populated=%d", got))}
	},
}
}

// profileMemoryDPC reports a different number of modules per channel.
func profileMemoryDPC() Rule { return Rule{
	ID:       "profile.memory.dimms_per_channel",
	Title:    "Modules per channel do not match the expected profile",
	Severity: Critical,
	Source:   "https://lenovopress.lenovo.com/lp2000-balanced-memory-configurations-4th-5th-gen-2s-intel-processors",
	Remedy:   "Repopulate the node or correct the profile.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.Memory == nil || p.Expect.Memory.DIMMsPerChannel == nil {
			return nil
		}
		var got int
		for _, ch := range c.Manifest.Memory.Channels {
			if ch.DIMMs > got {
				got = ch.DIMMs
			}
		}
		want := *p.Expect.Memory.DIMMsPerChannel
		if got == want {
			return nil
		}
		return []Finding{finding(profileMemoryDPC(),
			fmt.Sprintf("the busiest channel holds %d modules, expected %d", got, want),
			fmt.Sprintf("dimms_per_channel=%d", got))}
	},
}
}

// profileMemorySlots reports a different number of populated slots.
func profileMemorySlots() Rule { return Rule{
	ID:       "profile.memory.slots_filled",
	Title:    "Populated memory slots do not match the expected profile",
	Severity: Critical,
	Source:   "https://cray-hpe.github.io/docs-sat/en-25/introduction/",
	Remedy:   "Repopulate the node or correct the profile.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.Memory == nil || p.Expect.Memory.SlotsFilled == nil {
			return nil
		}
		got := c.Manifest.Memory.SlotsFilled
		want := *p.Expect.Memory.SlotsFilled
		if got == want {
			return nil
		}
		return []Finding{finding(profileMemorySlots(),
			fmt.Sprintf("%d memory slots are populated, expected %d", got, want),
			fmt.Sprintf("slots_filled=%d", got))}
	},
}
}

// profileMemorySpeed reports memory running at a speed other than the one
// the fleet expects.
func profileMemorySpeed() Rule { return Rule{
	ID:       "profile.memory.configured_mts",
	Title:    "Memory runs at a speed other than the expected profile",
	Severity: Critical,
	Source:   "https://server-configurator.com/learn/ddr5-server-memory-population-rules",
	Remedy:   "Fix the population or the firmware speed setting so the node matches the fleet.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.Memory == nil || p.Expect.Memory.ConfiguredMTs == nil {
			return nil
		}
		got := c.Manifest.Memory.ConfiguredMTs
		want := *p.Expect.Memory.ConfiguredMTs
		if got == 0 {
			return []Finding{finding(profileMemorySpeed(),
				fmt.Sprintf("the memory speed cannot be read, the profile expects %d MT/s", want),
				"SMBIOS table is not readable")}
		}
		if got == want {
			return nil
		}
		potential := collect.MemoryBandwidthGBs(len(c.Manifest.Memory.Channels), want)
		actual := collect.MemoryBandwidthGBs(len(c.Manifest.Memory.Channels), got)
		f := finding(profileMemorySpeed(),
			fmt.Sprintf("memory runs at %d MT/s, expected %d MT/s", got, want),
			fmt.Sprintf("configured_mts=%d", got),
			fmt.Sprintf("expected_mts=%d", want))
		f.Impact = &Impact{
			Kind:     "memory_bandwidth",
			LossPct:  lossPct(potential, actual),
			Potential: potential,
			Actual:   actual,
			Unit:     "GB/s",
		}
		return []Finding{f}
	},
}
}

// profileMemoryECC reports ECC state that differs from the profile.
func profileMemoryECC() Rule { return Rule{
	ID:       "profile.memory.ecc",
	Title:    "Memory error correction does not match the expected profile",
	Severity: Warn,
	Source:   "https://www.vmware.com/docs/vsphere-esxi-vcenter-server-9-1-performance-best-practices",
	Remedy:   "Enable ECC in firmware or correct the profile.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.Memory == nil || p.Expect.Memory.ECCEnabled == nil {
			return nil
		}
		got := c.Manifest.Memory.ECCEnabled
		want := *p.Expect.Memory.ECCEnabled
		if got == want {
			return nil
		}
		state := "disabled"
		if got {
			state = "enabled"
		}
		return []Finding{finding(profileMemoryECC(),
			fmt.Sprintf("memory error correction is %s, expected the opposite", state),
			fmt.Sprintf("ecc_enabled=%t", got))}
	},
}
}

// profilePCIeLink reports a link slower than the fleet expects.
func profilePCIeLink() Rule { return Rule{
	ID:       "profile.pcie.min_link",
	Title:    "A PCIe link is slower than the expected profile",
	Severity: Critical,
	Source:   "https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/_sources/troubleshooting/networking_troubleshooting.rst.txt",
	Remedy:   "Check the slot, the riser and the firmware link speed setting.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.PCIe == nil || p.Expect.PCIe.MinLinkGTs == nil {
			return nil
		}
		if p.Expect.PCIe.DegradedAllowed != nil && *p.Expect.PCIe.DegradedAllowed {
			return nil
		}
		min := *p.Expect.PCIe.MinLinkGTs
		var out []Finding
		for _, d := range c.Manifest.PCIe.Devices {
			if d.CurrentSpeedGTs <= 0 || d.CurrentSpeedGTs >= min {
				continue
			}
			out = append(out, finding(profilePCIeLink(),
				fmt.Sprintf("%s negotiated %.1f GT/s, the profile requires at least %.1f GT/s", d.Address, d.CurrentSpeedGTs, min),
				fmt.Sprintf("current_link_speed=%s", speedText(d.CurrentSpeedGTs))))
		}
		return out
	},
}
}

// profileOSConfig reports operating system settings that differ from the
// profile.
func profileOSConfig() Rule { return Rule{
	ID:       "profile.os.mismatch",
	Title:    "Operating system settings do not match the expected profile",
	Severity: Warn,
	Source:   "https://github.com/intel/PerfSpect",
	Remedy:   "Apply the fleet configuration to this node.",
	Check: func(c *Context) []Finding {
		p := c.Profile
		if p == nil || p.Expect.OS == nil {
			return nil
		}
		e := p.Expect.OS
		var out []Finding
		if e.TransparentHugepage != "" {
			if got := thpMode(c.Manifest.OS.THP); got != "" && !strings.EqualFold(got, e.TransparentHugepage) {
				out = append(out, finding(profileOSConfig(),
					fmt.Sprintf("transparent huge pages are %q, expected %q", got, e.TransparentHugepage),
					fmt.Sprintf("transparent_hugepage=%s", c.Manifest.OS.THP)))
			}
		}
		if e.Governor != "" && c.Manifest.CPU.Governor != "" &&
			!strings.EqualFold(e.Governor, c.Manifest.CPU.Governor) {
			out = append(out, finding(profileOSConfig(),
				fmt.Sprintf("the governor is %q, expected %q", c.Manifest.CPU.Governor, e.Governor),
				fmt.Sprintf("scaling_governor=%s", c.Manifest.CPU.Governor)))
		}
		if e.NumaBalancing != "" && c.Manifest.OS.NUMABalancing != "" &&
			!strings.EqualFold(e.NumaBalancing, c.Manifest.OS.NUMABalancing) {
			out = append(out, finding(profileOSConfig(),
				fmt.Sprintf("automatic NUMA balancing is %q, expected %q", c.Manifest.OS.NUMABalancing, e.NumaBalancing),
				fmt.Sprintf("numa_balancing=%s", c.Manifest.OS.NUMABalancing)))
		}
		return out
	},
}
}

// thpMode extracts the selected mode from the transparent hugepage string.
func thpMode(s string) string {
	if s == "" {
		return ""
	}
	if i := strings.Index(s, "["); i >= 0 {
		mode := s[i+1:]
		if j := strings.Index(mode, "]"); j >= 0 {
			mode = mode[:j]
		}
		return mode
	}
	return ""
}
