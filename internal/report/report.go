// Package report renders manifests, findings and diffs as plain text.
//
// The output is meant to be read in a terminal, pasted into a ticket, and
// parsed by a script. No colour, no curses, no paging.
package report

import (
	"fmt"
	"strings"

	"github.com/Balanov/asbuilt/internal/collect"
	"github.com/Balanov/asbuilt/internal/diff"
	"github.com/Balanov/asbuilt/internal/rules"
	"github.com/Balanov/asbuilt/internal/version"
)

// Manifest renders the machine summary.
func Manifest(m *collect.Manifest) string {
	var b strings.Builder

	fmt.Fprintf(&b, "ASBUILT %s  (manifest schema %s)\n", version.Version, m.Schema)
	fmt.Fprintf(&b, "%s  %s\n", m.Plat.Vendor, m.Plat.Product)
	if m.Plat.SKU != "" {
		fmt.Fprintf(&b, "SKU %s  serial %s\n", m.Plat.SKU, m.Plat.Serial)
	}
	if m.Plat.BIOS.Version != "" {
		fmt.Fprintf(&b, "BIOS  %s %s (%s)\n", m.Plat.BIOS.Vendor, m.Plat.BIOS.Version, m.Plat.BIOS.ReleaseDate)
	}
	if m.Plat.Board.Product != "" {
		fmt.Fprintf(&b, "BOARD %s %s\n", m.Plat.Board.Manufacturer, m.Plat.Board.Product)
	}
	b.WriteString("\n")

	cpu := m.CPU
	fmt.Fprintf(&b, "CPU   %s\n", cpu.Model)
	fmt.Fprintf(&b, "      %d socket(s) x %d cores, %d threads per core, %d logical CPUs, SMT %s\n",
		cpu.Sockets, cpu.CoresPerSocket, cpu.ThreadsPerCore, cpu.LogicalCPUs, onOff(cpu.SMTEnabled))
	if cpu.MaxFreqMHz > 0 {
		fmt.Fprintf(&b, "      %d MHz base, %d MHz max, turbo %s", cpu.BaseFreqMHz, cpu.MaxFreqMHz, onOff(cpu.TurboEnabled))
		if cpu.Governor != "" {
			fmt.Fprintf(&b, ", governor %s", cpu.Governor)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "      ISA fingerprint %s (%d features)\n", cpu.ISA.Fingerprint, cpu.ISA.Count)
	if cpu.Cache.L3KB > 0 {
		fmt.Fprintf(&b, "      cache L1d %dK, L1i %dK, L2 %dK, L3 %dK\n",
			cpu.Cache.L1dKB, cpu.Cache.L1iKB, cpu.Cache.L2KB, cpu.Cache.L3KB)
	}
	b.WriteString("\n")

	mem := m.Memory
	fmt.Fprintf(&b, "MEM   %d channels, %d DIMMs, %d MB total\n", len(mem.Channels), mem.SlotsFilled, mem.TotalKB/1024)
	if mem.ConfiguredMTs > 0 {
		fmt.Fprintf(&b, "      %d MT/s configured", mem.ConfiguredMTs)
		if mem.MaxMTs > mem.ConfiguredMTs {
			fmt.Fprintf(&b, ", modules rated %d MT/s", mem.MaxMTs)
		}
		b.WriteString("\n")
	}
	if mem.BandwidthGBs > 0 {
		fmt.Fprintf(&b, "      peak bandwidth %.1f GB/s", mem.BandwidthGBs)
		if mem.BandwidthAtMaxGBs > mem.BandwidthGBs {
			fmt.Fprintf(&b, " (%.1f GB/s at rated speed)", mem.BandwidthAtMaxGBs)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "      ECC %s, source %s\n", onOff(mem.ECCEnabled), mem.Source)
	b.WriteString("\n")

	fmt.Fprintf(&b, "NUMA  %d node(s)", m.NUMA.Nodes)
	if m.NUMA.Interleaved {
		b.WriteString(" (interleaved: the firmware hides the real topology)")
	}
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "PCIE  %d functions, %d degraded, AER errors %s\n",
		len(m.PCIe.Devices), m.PCIe.Degraded, onOff(m.PCIe.AERError))
	for _, d := range m.PCIe.Devices {
		if !d.Degraded {
			continue
		}
		fmt.Fprintf(&b, "      %s %.1f GT/s x%d of %.1f GT/s x%d\n",
			d.Address, d.CurrentSpeedGTs, d.CurrentWidth, d.MaxSpeedGTs, d.MaxWidth)
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "OS    %s\n", m.OS.Kernel)
	if m.OS.THP != "" {
		fmt.Fprintf(&b, "      transparent huge pages: %s\n", m.OS.THP)
	}
	if m.OS.ASPM != "" {
		fmt.Fprintf(&b, "      PCIe ASPM policy: %s\n", m.OS.ASPM)
	}
	if m.Power.ThrottleThermal > 0 || m.Power.ThrottlePower > 0 {
		fmt.Fprintf(&b, "      throttling: %d thermal, %d power limit events since boot\n",
			m.Power.ThrottleThermal, m.Power.ThrottlePower)
	}
	if m.Virt.Hypervisor != "" {
		fmt.Fprintf(&b, "      virtualization: %s\n", m.Virt.Hypervisor)
	}
	if m.Host.Containerized {
		fmt.Fprintf(&b, "      container: %s\n", m.Host.ContainerType)
	}
	if len(m.Notes) > 0 {
		b.WriteString("\nNOTES\n")
		for _, n := range m.Notes {
			fmt.Fprintf(&b, "  %s\n", n)
		}
	}
	return b.String()
}

// Findings renders the audit findings.
func Findings(findings []rules.Finding) string {
	var b strings.Builder
	if len(findings) == 0 {
		b.WriteString("FINDINGS  none: the machine matches every rule asbuilt checks.\n")
		return b.String()
	}
	counts := map[rules.Severity]int{}
	for _, f := range findings {
		counts[f.Severity]++
	}
	fmt.Fprintf(&b, "FINDINGS  %d critical, %d warn, %d info\n\n",
		counts[rules.Critical], counts[rules.Warn], counts[rules.Info])
	for _, f := range findings {
		fmt.Fprintf(&b, "%-8s %s\n", strings.ToUpper(string(f.Severity)), f.Title)
		fmt.Fprintf(&b, "         %s\n", f.ID)
		fmt.Fprintf(&b, "         %s\n", f.Detail)
		if f.Impact != nil {
			if f.Impact.LossPct > 0 {
				fmt.Fprintf(&b, "         impact: %s -%.1f%% (%.1f of %.1f %s)\n",
					f.Impact.Kind, f.Impact.LossPct, f.Impact.Actual, f.Impact.Potential, f.Impact.Unit)
			} else {
				fmt.Fprintf(&b, "         impact: %s\n", f.Impact.Kind)
			}
		}
		for _, e := range f.Evidence {
			fmt.Fprintf(&b, "         evidence: %s\n", e)
		}
		if f.Remedy != "" {
			fmt.Fprintf(&b, "         remedy: %s\n", f.Remedy)
		}
		if f.Source != "" {
			fmt.Fprintf(&b, "         source: %s\n", f.Source)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// Diff renders the differences between two manifests.
func Diff(diffs []diff.Difference) string {
	var b strings.Builder
	if len(diffs) == 0 {
		b.WriteString("DIFF  no differences: the two machines are identical as built.\n")
		return b.String()
	}
	fmt.Fprintf(&b, "DIFF  %d difference(s)\n\n", len(diffs))
	for _, d := range diffs {
		fmt.Fprintf(&b, "  %s\n", d.Path)
		fmt.Fprintf(&b, "    a: %s\n", d.A)
		fmt.Fprintf(&b, "    b: %s\n", d.B)
		if d.Impact != "" {
			fmt.Fprintf(&b, "    affects: %s\n", d.Impact)
		}
	}
	return b.String()
}

// Catalog renders the rule catalog.
func Catalog(catalog []rules.Rule) string {
	var b strings.Builder
	fmt.Fprintf(&b, "RULES  %d\n\n", len(catalog))
	for _, r := range catalog {
		fmt.Fprintf(&b, "%-8s %-42s %s\n", strings.ToUpper(string(r.Severity)), r.ID, r.Title)
	}
	return b.String()
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
