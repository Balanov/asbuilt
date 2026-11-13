package collect

import (
	"os"
	"strconv"
	"strings"

	"github.com/Balanov/asbuilt/internal/dmi"
	"github.com/Balanov/asbuilt/internal/sysfs"
)

// Tool is the name this collector writes into every manifest.
const Tool = "asbuilt"

// Gather collects the as-built truth of the machine behind fs.
//
// Gathering never fails because a source is missing: an unreachable source
// becomes a note and an empty section, because "we could not prove this" is
// itself a fact an operator needs.
func Gather(f *sysfs.FS) *Manifest {
	m := &Manifest{
		Schema:  SchemaVersion,
		Tool:    Tool,
		Sources: map[string]string{},
	}

	m.Host.Hostname = readString(f, "/etc/hostname")
	if m.Host.Hostname == "" {
		if h, err := os.Hostname(); err == nil {
			m.Host.Hostname = h
		}
	}
	if up, err := f.ReadString("/proc/uptime"); err == nil {
		fields := strings.Fields(up)
		if len(fields) > 0 {
			m.Host.UptimeSeconds = int64(parseFloat(fields[0]))
		}
	}
	m.Host.Containerized, m.Host.ContainerType = detectContainer(f)

	collectPlatform(f, m)
	collectCPU(f, m)
	collectMemory(f, m)
	collectPCIe(f, m)
	collectSystem(f, m)

	return m
}

// collectPlatform fills the firmware identity.
func collectPlatform(f *sysfs.FS, m *Manifest) {
	raw, err := f.ReadFile("/sys/firmware/dmi/tables/DMI")
	if err != nil {
		m.Notes = append(m.Notes, "platform: SMBIOS table is not readable (needs root)")
		return
	}
	table, err := dmi.Parse(raw)
	if err != nil {
		m.Notes = append(m.Notes, "platform: SMBIOS table could not be parsed: "+err.Error())
		return
	}
	m.Sources["smbios"] = "ok"

	if sys, ok := table.System(); ok {
		m.Plat.Vendor = sys.Manufacturer
		m.Plat.Product = sys.Product
		m.Plat.Version = sys.Version
		m.Plat.Serial = sys.Serial
		m.Plat.SKU = sys.SKU
		m.Plat.UUID = sys.UUID
	}
	if b, ok := table.BIOS(); ok {
		m.Plat.BIOS = BIOS{
			Vendor:      b.Vendor,
			Version:     b.Version,
			ReleaseDate: b.ReleaseDate,
		}
	}
	if bb, ok := table.Baseboard(); ok {
		m.Plat.Board = Board{
			Manufacturer: bb.Manufacturer,
			Product:      bb.Product,
			Version:      bb.Version,
		}
	}
}

// detectContainer reports whether the tool runs inside a container.
func detectContainer(f *sysfs.FS) (bool, string) {
	if f.Exists("/.dockerenv") {
		return true, "docker"
	}
	if f.Exists("/run/.containerenv") {
		return true, "podman"
	}
	if data, err := f.ReadFile("/proc/1/cgroup"); err == nil {
		cg := string(data)
		switch {
		case strings.Contains(cg, "docker"):
			return true, "docker"
		case strings.Contains(cg, "kubepods"):
			return true, "kubernetes"
		case strings.Contains(cg, "containerd"):
			return true, "containerd"
		case strings.Contains(cg, "lxc"):
			return true, "lxc"
		}
	}
	return false, ""
}

func parseFloat(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}
