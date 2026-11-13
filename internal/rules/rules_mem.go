package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Balanov/asbuilt/internal/collect"
)

// memSpeedDownclocked reports memory running below the rating of the
// modules themselves.
//
// The memory controller always runs the bus at the slowest bin it can
// safely use, so a downclock is a measurable loss of bandwidth rather than
// a matter of opinion. Whether the cause is the population, the mixed kit
// or the platform ceiling is a separate question, which the detail names.
func memSpeedDownclocked() Rule { return Rule{
	ID:       "mem.speed.below_module_rating",
	Title:    "Memory runs below the rating of the installed modules",
	Severity: Warn,
	Source:   "https://server-configurator.com/learn/ddr5-server-memory-population-rules",
	Remedy:   "Populate every channel symmetrically with identical modules, and check the platform speed table for this DIMM count per channel.",
	Check: func(c *Context) []Finding {
		var slow []collect.MemoryChannel
		for _, ch := range c.Manifest.Memory.Channels {
			if ch.ConfiguredMTs <= 0 || ch.MaxMTs <= 0 || ch.ConfiguredMTs >= ch.MaxMTs {
				continue
			}
			slow = append(slow, ch)
		}
		if len(slow) == 0 {
			return nil
		}
		first := slow[0]
		potential := collect.MemoryBandwidthGBs(len(c.Manifest.Memory.Channels), first.MaxMTs)
		actual := collect.MemoryBandwidthGBs(len(c.Manifest.Memory.Channels), first.ConfiguredMTs)
		ev := []string{
			fmt.Sprintf("channels below their module rating: %d of %d", len(slow), len(c.Manifest.Memory.Channels)),
		}
		for _, ch := range slow {
			ev = append(ev, fmt.Sprintf("channel %s: %d MT/s configured, %d MT/s rated", ch.ID, ch.ConfiguredMTs, ch.MaxMTs))
		}
		f := finding(memSpeedDownclocked(),
			fmt.Sprintf("memory runs at %d MT/s while the installed modules are rated for %d MT/s", first.ConfiguredMTs, first.MaxMTs),
			ev...)
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

// memMixedSpeedWithinChannel reports modules of different ratings inside
// one channel, where the controller has to fall back to the slowest one.
func memMixedSpeedWithinChannel() Rule { return Rule{
	ID:       "mem.speed.mixed_within_channel",
	Title:    "A memory channel mixes modules of different rated speed",
	Severity: Critical,
	Source:   "https://www.intel.com/content/www/us/en/support/articles/000088926.html",
	Remedy:   "Replace the channel with one identical part number.",
	Check: func(c *Context) []Finding {
		var out []Finding
		for _, ch := range c.Manifest.Memory.Channels {
			speeds := map[int]bool{}
			for _, d := range c.Manifest.Memory.DIMMs {
				if d.Channel == ch.ID && d.MaxSpeedMTs > 0 {
					speeds[d.MaxSpeedMTs] = true
				}
			}
			if len(speeds) < 2 {
				continue
			}
			list := make([]int, 0, len(speeds))
			for s := range speeds {
				list = append(list, s)
			}
			sort.Ints(list)
			out = append(out, finding(memMixedSpeedWithinChannel(),
				fmt.Sprintf("channel %s mixes modules rated %s MT/s", ch.ID, joinInts(list)),
				fmt.Sprintf("channel %s: %d DIMMs", ch.ID, ch.DIMMs)))
		}
		return out
	},
}
}

// memMixedPartsWithinChannel reports a channel built from mixed part
// numbers, which vendors warn about explicitly.
func memMixedPartsWithinChannel() Rule { return Rule{
	ID:       "mem.parts.mixed_within_channel",
	Title:    "A memory channel mixes module part numbers",
	Severity: Warn,
	Source:   "https://www.intel.com/content/www/us/en/support/articles/000088926.html",
	Remedy:   "Use one part number per channel so that the controller trains a single speed bin.",
	Check: func(c *Context) []Finding {
		var out []Finding
		for _, ch := range c.Manifest.Memory.Channels {
			if len(ch.PartNumbers) < 2 {
				continue
			}
			out = append(out, finding(memMixedPartsWithinChannel(),
				fmt.Sprintf("channel %s carries %d different part numbers: %s",
					ch.ID, len(ch.PartNumbers), strings.Join(ch.PartNumbers, ", ")),
				fmt.Sprintf("channel %s: %d DIMMs", ch.ID, ch.DIMMs)))
		}
		return out
	},
}
}

// memPopulationUneven reports channels that are not populated alike.
//
// Lenovo measured an unbalanced population falling to 13% of the bandwidth
// of a balanced one; Dell measured a 65% to 35% STREAM drop. This is the
// single most expensive configuration mistake on a server, and it is
// invisible to every tool that only counts modules.
func memPopulationUneven() Rule { return Rule{
	ID:       "mem.population.uneven",
	Title:    "Memory channels are not populated alike",
	Severity: Critical,
	Source:   "https://lenovopress.lenovo.com/lp2000-balanced-memory-configurations-4th-5th-gen-2s-intel-processors",
	Remedy:   "Populate every channel with the same number of identical modules.",
	Check: func(c *Context) []Finding {
		chans := c.Manifest.Memory.Channels
		if len(chans) < 2 {
			return nil
		}
		minD, maxD := chans[0].DIMMs, chans[0].DIMMs
		minMB, maxMB := chans[0].TotalMB, chans[0].TotalMB
		for _, ch := range chans[1:] {
			if ch.DIMMs < minD {
				minD = ch.DIMMs
			}
			if ch.DIMMs > maxD {
				maxD = ch.DIMMs
			}
			if ch.TotalMB < minMB {
				minMB = ch.TotalMB
			}
			if ch.TotalMB > maxMB {
				maxMB = ch.TotalMB
			}
		}
		if minD == maxD && minMB == maxMB {
			return nil
		}
		var ev []string
		for _, ch := range chans {
			ev = append(ev, fmt.Sprintf("channel %s: %d DIMMs, %d MB", ch.ID, ch.DIMMs, ch.TotalMB))
		}
		f := finding(memPopulationUneven(),
			fmt.Sprintf("channels hold between %d and %d DIMMs and between %d and %d MB", minD, maxD, minMB, maxMB),
			ev...)
		potential := collect.MemoryBandwidthGBs(len(chans), c.Manifest.Memory.ConfiguredMTs)
		actual := collect.MemoryBandwidthGBs(len(chans), c.Manifest.Memory.ConfiguredMTs) * ratio(minMB, maxMB)
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

// memSlotsEmpty reports slots the board offers but nobody filled.
func memSlotsEmpty() Rule { return Rule{
	ID:       "mem.slots.empty",
	Title:    "Memory slots are empty",
	Severity: Info,
	Source:   "https://server-configurator.com/learn/server-memory-channels-ranks",
	Remedy:   "Fill every channel before adding capacity, and keep both sockets symmetric.",
	Check: func(c *Context) []Finding {
		mem := c.Manifest.Memory
		if mem.SlotsTotal <= 0 || mem.SlotsFilled >= mem.SlotsTotal {
			return nil
		}
		empty := mem.SlotsTotal - mem.SlotsFilled
		return []Finding{finding(memSlotsEmpty(),
			fmt.Sprintf("%d of %d memory slots are empty", empty, mem.SlotsTotal),
			fmt.Sprintf("populated DIMMs: %d", mem.SlotsFilled),
			fmt.Sprintf("slots the board reports: %d", mem.SlotsTotal))}
	},
}
}

// memECCDisabled reports a server without memory error correction.
func memECCDisabled() Rule { return Rule{
	ID:       "mem.ecc.disabled",
	Title:    "No memory error correction is exposed",
	Severity: Warn,
	Source:   "https://www.vmware.com/docs/vsphere-esxi-vcenter-server-9-1-performance-best-practices",
	Remedy:   "Enable ECC in firmware; a server without it cannot detect a failing module.",
	Check: func(c *Context) []Finding {
		if c.Manifest.Memory.Source != "smbios" {
			return nil
		}
		if c.Manifest.Memory.ECCEnabled {
			return nil
		}
		return []Finding{finding(memECCDisabled(),
			"the kernel exposes no EDAC memory controller, so memory errors are not being corrected",
			"no /sys/devices/system/edac/mc entry")}
	},
}
}

// numaSingleNodeMultiSocket reports a multi-socket machine that presents a
// single memory node, which means the firmware interleaved memory and hid
// the real topology from the operating system.
func numaSingleNodeMultiSocket() Rule { return Rule{
	ID:       "numa.single_node_multi_socket",
	Title:    "Multi-socket machine presents a single NUMA node",
	Severity: Critical,
	Source:   "https://www.intel.com/content/www/us/en/developer/articles/technical/measuring-impact-of-numa-migrations-on-performance.html",
	Remedy:   "Disable node interleaving in firmware so that the topology is visible again.",
	Check: func(c *Context) []Finding {
		if !c.Manifest.NUMA.Interleaved {
			return nil
		}
		return []Finding{finding(numaSingleNodeMultiSocket(),
			fmt.Sprintf("%d sockets expose %d NUMA node: memory interleaving hides the real topology",
				c.Manifest.CPU.Sockets, c.Manifest.NUMA.Nodes),
			fmt.Sprintf("sockets: %d", c.Manifest.CPU.Sockets),
			fmt.Sprintf("NUMA nodes: %d", c.Manifest.NUMA.Nodes))}
	},
}
}

// numaAsymmetric reports memory nodes that do not hold the same amount of
// memory.
func numaAsymmetric() Rule { return Rule{
	ID:       "numa.asymmetric",
	Title:    "NUMA nodes hold different amounts of memory",
	Severity: Critical,
	Source:   "https://lenovopress.lenovo.com/lp2000-balanced-memory-configurations-4th-5th-gen-2s-intel-processors",
	Remedy:   "Populate the sockets symmetrically.",
	Check: func(c *Context) []Finding {
		nodes := c.Manifest.NUMA.Topology
		if len(nodes) < 2 {
			return nil
		}
		minKB, maxKB := nodes[0].MemTotalKB, nodes[0].MemTotalKB
		for _, n := range nodes[1:] {
			if n.MemTotalKB < minKB {
				minKB = n.MemTotalKB
			}
			if n.MemTotalKB > maxKB {
				maxKB = n.MemTotalKB
			}
		}
		if minKB == maxKB || maxKB == 0 {
			return nil
		}
		var ev []string
		for _, n := range nodes {
			ev = append(ev, fmt.Sprintf("node %d: %d MB", n.ID, n.MemTotalKB/1024))
		}
		return []Finding{finding(numaAsymmetric(),
			fmt.Sprintf("NUMA nodes hold between %d MB and %d MB", minKB/1024, maxKB/1024), ev...)}
	},
}
}

func joinInts(v []int) string {
	parts := make([]string, len(v))
	for i, n := range v {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(parts, "/")
}

func ratio(a, b int) float64 {
	if b <= 0 {
		return 1
	}
	return float64(a) / float64(b)
}
