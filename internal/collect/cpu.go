package collect

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Balanov/asbuilt/internal/sysfs"
)

// cpuInfo is the parsed subset of /proc/cpuinfo.
type cpuInfo struct {
	Vendor      string
	Model       string
	Family      string
	ModelNum    string
	Stepping    string
	Microcode   string
	Flags       string
	PhysicalIDs map[string]bool
	Siblings    int
	Cores       int
	Processors  int
}

func parseCPUInfo(text string) cpuInfo {
	ci := cpuInfo{PhysicalIDs: map[string]bool{}}
	for _, line := range strings.Split(text, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "processor":
			ci.Processors++
		case "vendor_id":
			if ci.Vendor == "" {
				ci.Vendor = value
			}
		case "model name":
			if ci.Model == "" {
				ci.Model = value
			}
		case "cpu family":
			if ci.Family == "" {
				ci.Family = value
			}
		case "model":
			if ci.ModelNum == "" {
				ci.ModelNum = value
			}
		case "stepping":
			if ci.Stepping == "" {
				ci.Stepping = value
			}
		case "microcode":
			if ci.Microcode == "" {
				ci.Microcode = value
			}
		case "flags":
			if ci.Flags == "" {
				ci.Flags = value
			}
		case "physical id":
			ci.PhysicalIDs[value] = true
		case "siblings":
			if n, err := strconv.Atoi(value); err == nil && n > ci.Siblings {
				ci.Siblings = n
			}
		case "cpu cores":
			if n, err := strconv.Atoi(value); err == nil && n > ci.Cores {
				ci.Cores = n
			}
		}
	}
	return ci
}

// collectCPU fills the CPU section of the manifest.
func collectCPU(f *sysfs.FS, m *Manifest) {
	raw, err := f.ReadFile("/proc/cpuinfo")
	if err != nil {
		m.Notes = append(m.Notes, "cpu: /proc/cpuinfo is not readable")
		return
	}
	ci := parseCPUInfo(string(raw))

	m.CPU.Vendor = ci.Vendor
	m.CPU.Model = ci.Model
	m.CPU.Family = ci.Family
	m.CPU.Stepping = ci.Stepping
	m.CPU.Microcode = ci.Microcode
	m.CPU.LogicalCPUs = ci.Processors
	m.CPU.Sockets = len(ci.PhysicalIDs)
	if m.CPU.Sockets == 0 && ci.Processors > 0 {
		m.CPU.Sockets = 1
	}
	m.CPU.CoresPerSocket = ci.Cores
	if ci.Siblings > 0 && ci.Cores > 0 {
		m.CPU.ThreadsPerCore = ci.Siblings / ci.Cores
	} else if ci.Processors > 0 && m.CPU.Sockets > 0 {
		m.CPU.ThreadsPerCore = ci.Processors / m.CPU.Sockets / max(ci.Cores, 1)
	}
	if m.CPU.ThreadsPerCore < 1 {
		m.CPU.ThreadsPerCore = 1
	}
	m.CPU.SMTEnabled = m.CPU.ThreadsPerCore > 1

	flags := strings.Fields(ci.Flags)
	sort.Strings(flags)
	m.CPU.ISA.Exposed = flags
	m.CPU.ISA.Count = len(flags)
	m.CPU.ISA.Fingerprint = fingerprint(flags)

	collectCPUCache(f, &m.CPU.Cache, m.CPU.LogicalCPUs)
	collectCPUFreq(f, &m.CPU)
}

// collectCPUCache reads the cache hierarchy of the first logical CPU.
func collectCPUCache(f *sysfs.FS, c *Cache, logicalCPUs int) {
	for _, name := range f.List("/sys/devices/system/cpu/cpu0/cache") {
		if !strings.HasPrefix(name, "index") {
			continue
		}
		base := "/sys/devices/system/cpu/cpu0/cache/" + name
		level, _ := f.ReadString(base + "/level")
		typ, _ := f.ReadString(base + "/type")
		size, _ := f.ReadString(base + "/size")
		shared, _ := f.ReadString(base + "/shared_cpu_list")
		kb := parseCacheSize(size)
		switch {
		case level == "1" && strings.EqualFold(typ, "Data"):
			c.L1dKB = kb
		case level == "1" && strings.EqualFold(typ, "Instruction"):
			c.L1iKB = kb
		case level == "2":
			c.L2KB = kb
		case level == "3":
			c.L3KB = kb
			if c.CacheMask == "" {
				c.CacheMask = shared
			}
		}
	}
	if c.L3KB > 0 && c.CacheMask != "" {
		c.L3Shared = logicalCPUs > 0 && countCPUList(c.CacheMask) == logicalCPUs
	}
}

func parseCacheSize(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	mult := 1
	switch {
	case strings.HasSuffix(s, "K"), strings.HasSuffix(s, "k"):
		mult = 1
		s = strings.TrimRight(s, "Kk")
	case strings.HasSuffix(s, "M"), strings.HasSuffix(s, "m"):
		mult = 1024
		s = strings.TrimRight(s, "Mm")
	case strings.HasSuffix(s, "G"), strings.HasSuffix(s, "g"):
		mult = 1024 * 1024
		s = strings.TrimRight(s, "Gg")
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n * mult
}

// collectCPUFreq reads the cpufreq policy of the first logical CPU.
func collectCPUFreq(f *sysfs.FS, c *CPU) {
	base := "/sys/devices/system/cpu/cpu0/cpufreq"
	if gov, err := f.ReadString(base + "/scaling_governor"); err == nil {
		c.Governor = gov
	}
	if khz, err := f.ReadString(base + "/base_frequency"); err == nil {
		c.BaseFreqMHz = atoiSafe(khz) / 1000
	}
	if khz, err := f.ReadString(base + "/scaling_max_freq"); err == nil {
		c.MaxFreqMHz = atoiSafe(khz) / 1000
	}
	if c.BaseFreqMHz == 0 {
		if khz, err := f.ReadString(base + "/cpuinfo_max_freq"); err == nil {
			c.MaxFreqMHz = atoiSafe(khz) / 1000
		}
	}
	if c.BaseFreqMHz > 0 && c.MaxFreqMHz > 0 {
		c.TurboEnabled = c.MaxFreqMHz > c.BaseFreqMHz
	}
}

func countCPUList(list string) int {
	total := 0
	for _, part := range strings.Split(list, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if lo, hi, ok := strings.Cut(part, "-"); ok {
			total += atoiSafe(hi) - atoiSafe(lo) + 1
		} else {
			total++
		}
	}
	return total
}

func atoiSafe(s string) int {
	s = strings.TrimSpace(s)
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func fingerprint(parts []string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, ",")))
	return hex.EncodeToString(sum[:])[:16]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func fmtMHz(mhz int) string { return fmt.Sprintf("%d MHz", mhz) }
