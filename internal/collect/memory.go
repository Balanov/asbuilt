package collect

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Balanov/asbuilt/internal/dmi"
	"github.com/Balanov/asbuilt/internal/sysfs"
)

// collectMemory fills the memory section.
//
// The firmware SMBIOS table is the only source that says which modules are
// actually populated, how many channels the board has, and at which speed
// the memory controller is really running. It is root-only on Linux, so the
// collector degrades to what /proc and /sys can prove and says so.
func collectMemory(f *sysfs.FS, m *Manifest) {
	if raw, err := f.ReadFile("/proc/meminfo"); err == nil {
		m.Memory.TotalKB = memTotalKB(string(raw))
	}

	raw, err := f.ReadFile("/sys/firmware/dmi/tables/DMI")
	if err != nil {
		m.Memory.Source = "sysfs"
		m.Notes = append(m.Notes, "memory: SMBIOS table is not readable (needs root); DIMM population is unknown")
		m.Memory.BandwidthGBs = 0
		return
	}
	table, err := dmi.Parse(raw)
	if err != nil {
		m.Memory.Source = "sysfs"
		m.Notes = append(m.Notes, "memory: SMBIOS table could not be parsed: "+err.Error())
		return
	}
	m.Memory.Source = "smbios"

	devices := table.MemoryDevices()
	arrays := table.MemoryArrays()

	for _, a := range arrays {
		m.Memory.SlotsTotal += a.DeviceCount
	}

	byChannel := map[string][]dmi.MemoryDevice{}
	order := []string{}
	for _, d := range devices {
		ch := dmi.ChannelOf(d.Locator)
		if ch == "" {
			ch = "unknown"
		}
		if _, seen := byChannel[ch]; !seen {
			order = append(order, ch)
		}
		byChannel[ch] = append(byChannel[ch], d)
	}
	sort.Strings(order)

	for _, ch := range order {
		mods := byChannel[ch]
		c := MemoryChannel{ID: ch, DIMMs: len(mods)}
		seenPN := map[string]bool{}
		for _, d := range mods {
			c.TotalMB += d.SizeMB
			if d.ConfiguredSpeedMTs > c.ConfiguredMTs {
				c.ConfiguredMTs = d.ConfiguredSpeedMTs
			}
			if d.SpeedMTs > c.MaxMTs {
				c.MaxMTs = d.SpeedMTs
			}
			if d.PartNumber != "" && !seenPN[d.PartNumber] {
				seenPN[d.PartNumber] = true
				c.PartNumbers = append(c.PartNumbers, d.PartNumber)
			}
			m.Memory.DIMMs = append(m.Memory.DIMMs, MemoryDIMM{
				Locator:            d.Locator,
				Channel:            ch,
				Socket:             dmi.SocketOf(d.Locator),
				SizeMB:             d.SizeMB,
				Type:               d.MemoryType,
				Vendor:             d.Vendor,
				PartNumber:         d.PartNumber,
				RankCount:          d.RankCount,
				ConfiguredSpeedMTs: d.ConfiguredSpeedMTs,
				MaxSpeedMTs:        d.SpeedMTs,
			})
		}
		m.Memory.Channels = append(m.Memory.Channels, c)
		m.Memory.SlotsFilled += len(mods)
	}

	// The effective data rate is the slowest populated channel: one slow
	// module drags the whole controller to its own bin.
	for _, c := range m.Memory.Channels {
		if c.ConfiguredMTs == 0 {
			continue
		}
		if m.Memory.ConfiguredMTs == 0 || c.ConfiguredMTs < m.Memory.ConfiguredMTs {
			m.Memory.ConfiguredMTs = c.ConfiguredMTs
		}
		if c.MaxMTs > m.Memory.MaxMTs {
			m.Memory.MaxMTs = c.MaxMTs
		}
	}

	m.Memory.BandwidthGBs = MemoryBandwidthGBs(len(m.Memory.Channels), m.Memory.ConfiguredMTs)
	m.Memory.BandwidthAtMaxGBs = MemoryBandwidthGBs(len(m.Memory.Channels), m.Memory.MaxMTs)
	m.Memory.ECCEnabled = detectECC(f)
	for _, a := range arrays {
		if strings.Contains(a.ErrorCorrection, "ecc") {
			m.Memory.ECCEnabled = true
		}
	}
}

// MemoryBandwidthGBs returns the theoretical peak of the given channels.
//
// Each channel moves 8 bytes per transfer on a 64-bit bus, so the peak is
// channels x data rate x 8 bytes per second.
func MemoryBandwidthGBs(channels, mts int) float64 {
	if channels <= 0 || mts <= 0 {
		return 0
	}
	return float64(channels) * float64(mts) * 8 / 1000
}

// detectECC reports whether the machine exposes memory error correction.
func detectECC(f *sysfs.FS) bool {
	for _, mc := range f.List("/sys/devices/system/edac/mc") {
		if strings.HasPrefix(mc, "mc") {
			for _, d := range f.List("/sys/devices/system/edac/mc/" + mc) {
				if strings.HasPrefix(d, "dimm") {
					return true
				}
			}
		}
	}
	return false
}

func memTotalKB(text string) int64 {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				n, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}
