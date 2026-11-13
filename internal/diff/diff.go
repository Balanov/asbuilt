// Package diff compares two manifests.
//
// The diff is what turns a pile of JSON into an audit: two nodes that were
// built the same way should produce an empty diff, and every line that does
// appear is either a configuration mistake or a hardware defect.
package diff

import (
	"fmt"

	"github.com/Balanov/asbuilt/internal/collect"
)

// Difference is one disagreement between two manifests.
type Difference struct {
	Path   string `json:"path"`
	A      string `json:"a"`
	B      string `json:"b"`
	Kind   string `json:"kind"`
	Impact string `json:"impact,omitempty"`
}

// Manifests returns every difference between two manifests.
//
// Only identity-bearing fields are compared: counters that move with every
// workload would drown the real differences.
func Manifests(a, b *collect.Manifest) []Difference {
	var out []Difference
	add := func(path, av, bv, impact string) {
		if av != bv {
			out = append(out, Difference{Path: path, A: av, B: bv, Kind: "value", Impact: impact})
		}
	}

	add("platform.vendor", a.Plat.Vendor, b.Plat.Vendor, "")
	add("platform.product", a.Plat.Product, b.Plat.Product, "")
	add("platform.version", a.Plat.Version, b.Plat.Version, "")
	add("platform.serial", a.Plat.Serial, b.Plat.Serial, "")
	add("platform.sku", a.Plat.SKU, b.Plat.SKU, "")
	add("platform.uuid", a.Plat.UUID, b.Plat.UUID, "")
	add("platform.bios.vendor", a.Plat.BIOS.Vendor, b.Plat.BIOS.Vendor, "")
	add("platform.bios.version", a.Plat.BIOS.Version, b.Plat.BIOS.Version, "")
	add("platform.bios.release_date", a.Plat.BIOS.ReleaseDate, b.Plat.BIOS.ReleaseDate, "")
	add("platform.board.product", a.Plat.Board.Product, b.Plat.Board.Product, "")

	add("cpu.model", a.CPU.Model, b.CPU.Model, "")
	add("cpu.microcode", a.CPU.Microcode, b.CPU.Microcode, "")
	add("cpu.sockets", itoa(a.CPU.Sockets), itoa(b.CPU.Sockets), "")
	add("cpu.cores_per_socket", itoa(a.CPU.CoresPerSocket), itoa(b.CPU.CoresPerSocket), "")
	add("cpu.threads_per_core", itoa(a.CPU.ThreadsPerCore), itoa(b.CPU.ThreadsPerCore), "")
	add("cpu.logical_cpus", itoa(a.CPU.LogicalCPUs), itoa(b.CPU.LogicalCPUs), "")
	add("cpu.smt_enabled", btoa(a.CPU.SMTEnabled), btoa(b.CPU.SMTEnabled), "")
	add("cpu.turbo_enabled", btoa(a.CPU.TurboEnabled), btoa(b.CPU.TurboEnabled), "")
	add("cpu.governor", a.CPU.Governor, b.CPU.Governor, "")
	add("cpu.isa.fingerprint", a.CPU.ISA.Fingerprint, b.CPU.ISA.Fingerprint, "")

	add("numa.nodes", itoa(a.NUMA.Nodes), itoa(b.NUMA.Nodes), "")

	add("memory.total_kb", itoa64(a.Memory.TotalKB), itoa64(b.Memory.TotalKB), "")
	add("memory.slots_filled", itoa(a.Memory.SlotsFilled), itoa(b.Memory.SlotsFilled), "")
	add("memory.configured_mts", itoa(a.Memory.ConfiguredMTs), itoa(b.Memory.ConfiguredMTs), "memory bandwidth")
	add("memory.max_mts", itoa(a.Memory.MaxMTs), itoa(b.Memory.MaxMTs), "memory bandwidth")
	add("memory.ecc_enabled", btoa(a.Memory.ECCEnabled), btoa(b.Memory.ECCEnabled), "")

	diffChannels(&out, a, b)
	diffDIMMMap(&out, a, b)
	diffPCIe(&out, a, b)

	add("os.kernel", a.OS.Kernel, b.OS.Kernel, "")
	add("os.transparent_hugepage", a.OS.THP, b.OS.THP, "")
	add("os.numa_balancing", a.OS.NUMABalancing, b.OS.NUMABalancing, "")
	add("os.pcie_aspm_policy", a.OS.ASPM, b.OS.ASPM, "")
	add("os.energy_perf_bias", a.OS.EnergyPerfBias, b.OS.EnergyPerfBias, "")

	return out
}

func diffChannels(out *[]Difference, a, b *collect.Manifest) {
	am := map[string]collect.MemoryChannel{}
	for _, ch := range a.Memory.Channels {
		am[ch.ID] = ch
	}
	bm := map[string]collect.MemoryChannel{}
	for _, ch := range b.Memory.Channels {
		bm[ch.ID] = ch
	}
	for _, id := range sortedKeys(am, bm) {
		ach, aok := am[id]
		bch, bok := bm[id]
		switch {
		case aok && !bok:
			*out = append(*out, Difference{Path: "memory.channels." + id, A: describeChannel(ach), B: "absent", Kind: "only_in_a"})
		case !aok && bok:
			*out = append(*out, Difference{Path: "memory.channels." + id, A: "absent", B: describeChannel(bch), Kind: "only_in_b"})
		default:
			*out = append(*out, compare("memory.channels."+id+".dimms", itoa(ach.DIMMs), itoa(bch.DIMMs), "memory bandwidth")...)
			*out = append(*out, compare("memory.channels."+id+".total_mb", itoa(ach.TotalMB), itoa(bch.TotalMB), "memory bandwidth")...)
			*out = append(*out, compare("memory.channels."+id+".configured_mts", itoa(ach.ConfiguredMTs), itoa(bch.ConfiguredMTs), "memory bandwidth")...)
		}
	}
}

func diffDIMMMap(out *[]Difference, a, b *collect.Manifest) {
	am := map[string]collect.MemoryDIMM{}
	for _, d := range a.Memory.DIMMs {
		am[d.Locator] = d
	}
	bm := map[string]collect.MemoryDIMM{}
	for _, d := range b.Memory.DIMMs {
		bm[d.Locator] = d
	}
	for _, loc := range sortedKeys(am, bm) {
		ad, aok := am[loc]
		bd, bok := bm[loc]
		switch {
		case aok && !bok:
			*out = append(*out, Difference{Path: "memory.dimms." + loc, A: describeDIMM(ad), B: "absent", Kind: "only_in_a"})
		case !aok && bok:
			*out = append(*out, Difference{Path: "memory.dimms." + loc, A: "absent", B: describeDIMM(bd), Kind: "only_in_b"})
		default:
			*out = append(*out, compare("memory.dimms."+loc+".size_mb", itoa(ad.SizeMB), itoa(bd.SizeMB), "")...)
			*out = append(*out, compare("memory.dimms."+loc+".part_number", ad.PartNumber, bd.PartNumber, "")...)
			*out = append(*out, compare("memory.dimms."+loc+".configured_speed_mts", itoa(ad.ConfiguredSpeedMTs), itoa(bd.ConfiguredSpeedMTs), "memory bandwidth")...)
		}
	}
}

func diffPCIe(out *[]Difference, a, b *collect.Manifest) {
	am := map[string]collect.PCIeDevice{}
	for _, d := range a.PCIe.Devices {
		am[d.Address] = d
	}
	bm := map[string]collect.PCIeDevice{}
	for _, d := range b.PCIe.Devices {
		bm[d.Address] = d
	}
	for _, addr := range sortedKeys(am, bm) {
		ad, aok := am[addr]
		bd, bok := bm[addr]
		switch {
		case aok && !bok:
			*out = append(*out, Difference{Path: "pcie." + addr, A: describePCIe(ad), B: "absent", Kind: "only_in_a"})
		case !aok && bok:
			*out = append(*out, Difference{Path: "pcie." + addr, A: "absent", B: describePCIe(bd), Kind: "only_in_b"})
		default:
			*out = append(*out, compare("pcie."+addr+".current_speed_gts", ftoa(ad.CurrentSpeedGTs), ftoa(bd.CurrentSpeedGTs), "pcie bandwidth")...)
			*out = append(*out, compare("pcie."+addr+".current_width", itoa(ad.CurrentWidth), itoa(bd.CurrentWidth), "pcie bandwidth")...)
			*out = append(*out, compare("pcie."+addr+".max_speed_gts", ftoa(ad.MaxSpeedGTs), ftoa(bd.MaxSpeedGTs), "pcie bandwidth")...)
			*out = append(*out, compare("pcie."+addr+".max_width", itoa(ad.MaxWidth), itoa(bd.MaxWidth), "pcie bandwidth")...)
		}
	}
}

func compare(path, av, bv, impact string) []Difference {
	if av == bv {
		return nil
	}
	return []Difference{{Path: path, A: av, B: bv, Kind: "value", Impact: impact}}
}

func describeChannel(c collect.MemoryChannel) string {
	return fmt.Sprintf("%d DIMMs, %d MB, %d MT/s", c.DIMMs, c.TotalMB, c.ConfiguredMTs)
}

func describeDIMM(d collect.MemoryDIMM) string {
	return fmt.Sprintf("%d MB, %s, %d MT/s", d.SizeMB, d.PartNumber, d.ConfiguredSpeedMTs)
}

func describePCIe(d collect.PCIeDevice) string {
	return fmt.Sprintf("%.1f GT/s x%d", d.CurrentSpeedGTs, d.CurrentWidth)
}

func sortedKeys[T any](a, b map[string]T) []string {
	seen := map[string]bool{}
	var keys []string
	for k := range a {
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	for k := range b {
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func itoa(v int) string      { return fmt.Sprintf("%d", v) }
func itoa64(v int64) string  { return fmt.Sprintf("%d", v) }
func btoa(v bool) string     { return fmt.Sprintf("%t", v) }
func ftoa(v float64) string  { return fmt.Sprintf("%.1f", v) }
