package collect

import (
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Balanov/asbuilt/internal/sysfs"
)

// pcieSpeedRe extracts the GT/s figure from a sysfs link speed string
// such as "16.0 GT/s PCIe".
var pcieSpeedRe = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*GT/s`)

// collectPCIe reads the state of every PCI function.
//
// The kernel publishes the negotiated link state as world-readable sysfs
// attributes, so the degraded-link check works without root and inside
// containers, where the PCI config space that lspci needs is not reachable.
func collectPCIe(f *sysfs.FS, m *Manifest) {
	const base = "/sys/bus/pci/devices"
	names := f.List(base)
	sort.Strings(names)

	// First pass: collect every function so that a device can be compared
	// against the upstream port that sits above it.
	byAddr := map[string]*PCIeDevice{}
	for _, name := range names {
		dev := &PCIeDevice{Address: name}
		dev.Vendor = readHex(f, base+"/"+name+"/vendor")
		dev.Device = readHex(f, base+"/"+name+"/device")
		dev.Class = readHex(f, base+"/"+name+"/class")
		dev.Driver = readDriver(f, base+"/"+name+"/driver")
		if n, err := f.ReadString(base + "/" + name + "/numa_node"); err == nil {
			dev.NUMANode, _ = strconv.Atoi(strings.TrimSpace(n))
		}
		dev.CurrentSpeedGTs = parseGTs(readString(f, base+"/"+name+"/current_link_speed"))
		dev.MaxSpeedGTs = parseGTs(readString(f, base+"/"+name+"/max_link_speed"))
		dev.CurrentWidth = atoiSafe(readString(f, base+"/"+name+"/current_link_width"))
		dev.MaxWidth = atoiSafe(readString(f, base+"/"+name+"/max_link_width"))
		dev.AERCorrectable = parseUint(readString(f, base+"/"+name+"/aer_dev_correctable"))
		dev.AERFatal = parseUint(readString(f, base+"/"+name+"/aer_dev_fatal")) +
			parseUint(readString(f, base+"/"+name+"/aer_dev_nonfatal"))
		byAddr[name] = dev
	}

	// Second pass: attach the upstream port limits. The sysfs device tree
	// encodes the hierarchy: a function lives in the directory of the
	// bridge above it, so the parent of a function is the nearest ancestor
	// directory that also names a PCI function.
	parents := pciParents(f)
	for name, dev := range byAddr {
		parent := parents[name]
		if parent == "" {
			continue
		}
		if p, ok := byAddr[parent]; ok {
			dev.PortMaxSpeedGTs = p.MaxSpeedGTs
			dev.PortMaxWidth = p.MaxWidth
		}
	}

	for _, name := range names {
		dev := byAddr[name]
		if dev.MaxSpeedGTs > 0 && dev.CurrentSpeedGTs > 0 {
			dev.LinkSpeedRatio = dev.CurrentSpeedGTs / dev.MaxSpeedGTs
		} else {
			dev.LinkSpeedRatio = 1
		}
		if dev.PortMaxSpeedGTs > 0 && dev.CurrentSpeedGTs > 0 {
			dev.PortSpeedRatio = dev.CurrentSpeedGTs / dev.PortMaxSpeedGTs
		} else {
			dev.PortSpeedRatio = 1
		}
		degraded := false
		if dev.MaxSpeedGTs > 0 && dev.CurrentSpeedGTs > 0 && dev.CurrentSpeedGTs < dev.MaxSpeedGTs {
			degraded = true
		}
		if dev.MaxWidth > 0 && dev.CurrentWidth > 0 && dev.CurrentWidth < dev.MaxWidth {
			degraded = true
		}
		dev.Degraded = degraded
		if degraded {
			m.PCIe.Degraded++
		}
		if dev.AERCorrectable > 0 || dev.AERFatal > 0 {
			m.PCIe.AERError = true
		}
		m.PCIe.Devices = append(m.PCIe.Devices, *dev)
	}
}

// pciParents maps every PCI function to the bridge directly above it.
//
// The kernel exposes the bus hierarchy as directories under
// /sys/devices/pci<domain>:<bus>. Walking that tree is enough, and it works
// without following symbolic links.
func pciParents(f *sysfs.FS) map[string]string {
	parents := map[string]string{}
	for _, root := range f.List("/sys/devices") {
		if !strings.HasPrefix(root, "pci") {
			continue
		}
		walkPCIeTree(f, "/sys/devices/"+root, "", parents)
	}
	return parents
}

// walkPCIeTree records the parent of every function found under dir.
func walkPCIeTree(f *sysfs.FS, dir, parent string, parents map[string]string) {
	for _, name := range f.List(dir) {
		if !isPCIAddress(name) {
			continue
		}
		child := dir + "/" + name
		if parent != "" {
			parents[name] = parent
		}
		walkPCIeTree(f, child, name, parents)
	}
}

// isPCIAddress reports whether a directory name looks like "0000:00:1c.0".
func isPCIAddress(name string) bool {
	if len(name) < 7 || strings.Count(name, ":") != 2 || !strings.Contains(name, ".") {
		return false
	}
	return true
}

// parseGTs extracts the data rate from a link speed string.
func parseGTs(s string) float64 {
	m := pcieSpeedRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return 0
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}
	return v
}

// LinkBandwidthGBs returns the per-direction bandwidth of a PCIe link.
//
// Generations 1-3 use 8b/10b line coding, generations 4 and later use
// 128b/130b. One lane carries one bit per GT/s, and eight bits make a byte.
func LinkBandwidthGBs(gts float64, width int) float64 {
	if gts <= 0 || width <= 0 {
		return 0
	}
	encoding := 0.8
	if gts >= 16 {
		encoding = 128.0 / 130.0
	}
	return gts * float64(width) * encoding / 8
}

func readString(f *sysfs.FS, p string) string {
	s, err := f.ReadString(p)
	if err != nil {
		return ""
	}
	return s
}

func readHex(f *sysfs.FS, p string) string {
	s := readString(f, p)
	if s == "" {
		return ""
	}
	return "0x" + strings.TrimPrefix(strings.TrimSpace(s), "0x")
}

func readDriver(f *sysfs.FS, p string) string {
	target, err := f.Readlink(p)
	if err != nil {
		return ""
	}
	base := path.Base(strings.TrimSuffix(target, "/"))
	return base
}

func parseUint(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
