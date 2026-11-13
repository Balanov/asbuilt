// Package collect gathers the as-built truth of a machine.
//
// Every collector is read-only, works without root where the kernel exposes
// the data unprivileged, and degrades explicitly when a source is missing.
// A missing source is a fact about the machine, not an error to hide.
package collect

// SchemaVersion is the version of the canonical manifest format.
// It changes only when the meaning of a field changes.
const SchemaVersion = "1"

// Manifest is the canonical description of a machine.
//
// The JSON encoding of this struct is the stable contract: it is what
// `asbuilt manifest` prints, what `asbuilt diff` compares, and what a
// profile is written against.
type Manifest struct {
	Schema string   `json:"schema"`
	Tool   string   `json:"tool"`
	Host   Host     `json:"host"`
	Plat   Platform `json:"platform"`
	CPU    CPU      `json:"cpu"`
	NUMA   NUMA     `json:"numa"`
	Memory Memory   `json:"memory"`
	PCIe   PCIe     `json:"pcie"`
	OS     OS       `json:"os"`
	Power  Power    `json:"power"`
	Virt   Virt     `json:"virtualization"`
	// Sources lists which data sources were reachable.
	Sources map[string]string `json:"sources,omitempty"`
	// Notes carries human-readable statements about missing evidence.
	Notes []string `json:"notes,omitempty"`
}

// Host identifies the running system.
type Host struct {
	Hostname      string `json:"hostname"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	Containerized bool   `json:"containerized"`
	ContainerType string `json:"container_type,omitempty"`
}

// Platform is the firmware-level identity of the machine.
type Platform struct {
	Vendor  string `json:"vendor,omitempty"`
	Product string `json:"product,omitempty"`
	Version string `json:"version,omitempty"`
	Serial  string `json:"serial,omitempty"`
	SKU     string `json:"sku,omitempty"`
	UUID    string `json:"uuid,omitempty"`
	BIOS    BIOS   `json:"bios"`
	Board   Board  `json:"board"`
	Chassis string `json:"chassis,omitempty"`
}

// BIOS is the firmware revision.
type BIOS struct {
	Vendor      string `json:"vendor,omitempty"`
	Version     string `json:"version,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
}

// Board is the mainboard identity.
type Board struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Product      string `json:"product,omitempty"`
	Version      string `json:"version,omitempty"`
}

// CPU describes the processors.
type CPU struct {
	Vendor         string `json:"vendor,omitempty"`
	Model          string `json:"model,omitempty"`
	Family         string `json:"family,omitempty"`
	Stepping       string `json:"stepping,omitempty"`
	Microcode      string `json:"microcode,omitempty"`
	Sockets        int    `json:"sockets"`
	CoresPerSocket int    `json:"cores_per_socket"`
	ThreadsPerCore int    `json:"threads_per_core"`
	LogicalCPUs    int    `json:"logical_cpus"`
	SMTEnabled     bool   `json:"smt_enabled"`
	BaseFreqMHz    int    `json:"base_freq_mhz,omitempty"`
	MaxFreqMHz     int    `json:"max_freq_mhz,omitempty"`
	TurboEnabled   bool   `json:"turbo_enabled"`
	Governor       string `json:"governor,omitempty"`
	Cache          Cache  `json:"cache"`
	ISA            ISA    `json:"isa"`
}

// Cache is the cache hierarchy of one socket.
type Cache struct {
	L1dKB     int  `json:"l1d_kb,omitempty"`
	L1iKB     int  `json:"l1i_kb,omitempty"`
	L2KB      int  `json:"l2kb,omitempty"`
	L3KB      int  `json:"l3kb,omitempty"`
	L3Shared  bool `json:"l3_shared"`
	CacheMask string `json:"cache_mask,omitempty"`
}

// ISA is the instruction set the operating system exposes.
type ISA struct {
	// Exposed is the sorted set of feature flags the kernel reports.
	Exposed []string `json:"exposed,omitempty"`
	// Fingerprint is a stable hash of Exposed, for fleet comparison.
	Fingerprint string `json:"fingerprint"`
	// Count is the number of exposed features.
	Count int `json:"count"`
}

// NUMA is the memory topology the kernel sees.
type NUMA struct {
	Nodes          int        `json:"nodes"`
	Topology       []NUMANode `json:"topology,omitempty"`
	Interleaved    bool       `json:"interleaved"`
	AutoBalancing  string     `json:"auto_balancing,omitempty"`
}

// NUMANode is one memory node.
type NUMANode struct {
	ID         int    `json:"id"`
	CPUs       string `json:"cpus,omitempty"`
	MemTotalKB int64  `json:"mem_total_kb"`
}

// Memory is the as-built memory configuration.
type Memory struct {
	TotalKB      int64           `json:"total_kb"`
	Source       string          `json:"source"`
	SlotsTotal   int             `json:"slots_total,omitempty"`
	SlotsFilled  int             `json:"slots_filled"`
	Channels     []MemoryChannel `json:"channels,omitempty"`
	DIMMs        []MemoryDIMM    `json:"dimms,omitempty"`
	ECCEnabled   bool            `json:"ecc_enabled"`
	ConfiguredMTs int            `json:"configured_mts,omitempty"`
	MaxMTs       int             `json:"max_mts,omitempty"`
	// BandwidthGBs is the theoretical peak from the as-built population.
	BandwidthGBs float64 `json:"bandwidth_gbs"`
	// BandwidthAtMaxGBs is the same figure if every module ran at its own
	// rated speed on a fully populated set of channels.
	BandwidthAtMaxGBs float64 `json:"bandwidth_at_max_gbs"`
}

// MemoryChannel is one memory channel.
type MemoryChannel struct {
	ID               string `json:"id"`
	DIMMs            int    `json:"dimms"`
	TotalMB          int    `json:"total_mb"`
	ConfiguredMTs    int    `json:"configured_mts,omitempty"`
	MaxMTs           int    `json:"max_mts,omitempty"`
	PartNumbers      []string `json:"part_numbers,omitempty"`
}

// MemoryDIMM is one populated module.
type MemoryDIMM struct {
	Locator          string `json:"locator"`
	Channel          string `json:"channel,omitempty"`
	Socket           string `json:"socket,omitempty"`
	SizeMB           int    `json:"size_mb"`
	Type             string `json:"type,omitempty"`
	Vendor           string `json:"vendor,omitempty"`
	PartNumber       string `json:"part_number,omitempty"`
	RankCount        int    `json:"rank_count,omitempty"`
	ConfiguredSpeedMTs int  `json:"configured_speed_mts,omitempty"`
	MaxSpeedMTs      int    `json:"max_speed_mts,omitempty"`
}

// PCIe holds the state of every PCI function found on the bus.
type PCIe struct {
	Devices  []PCIeDevice `json:"devices"`
	Degraded int          `json:"degraded"`
	AERError bool         `json:"aer_errors"`
}

// PCIeDevice is one PCI function.
type PCIeDevice struct {
	Address   string `json:"address"`
	Vendor    string `json:"vendor,omitempty"`
	Device    string `json:"device,omitempty"`
	Class     string `json:"class,omitempty"`
	Driver    string `json:"driver,omitempty"`
	NUMANode  int    `json:"numa_node"`
	// Link state: current is what the link negotiated, max is what the
	// endpoint is capable of, portMax is what the upstream port allows.
	CurrentSpeedGTs float64 `json:"current_speed_gts,omitempty"`
	CurrentWidth    int     `json:"current_width"`
	MaxSpeedGTs     float64 `json:"max_speed_gts,omitempty"`
	MaxWidth        int     `json:"max_width"`
	PortMaxSpeedGTs float64 `json:"port_max_speed_gts,omitempty"`
	PortMaxWidth    int     `json:"port_max_width"`
	// LinkSpeedRatio is current/max, PortSpeedRatio is current/portMax.
	LinkSpeedRatio float64 `json:"link_speed_ratio"`
	PortSpeedRatio float64 `json:"port_speed_ratio"`
	Degraded       bool    `json:"degraded"`
	AERCorrectable uint64  `json:"aer_correctable"`
	AERFatal       uint64  `json:"aer_fatal"`
}

// OS is the operating system view of the machine.
type OS struct {
	Kernel          string            `json:"kernel,omitempty"`
	Arch            string            `json:"arch,omitempty"`
	THP             string            `json:"transparent_hugepage,omitempty"`
	NUMABalancing   string            `json:"numa_balancing,omitempty"`
	ASPM            string            `json:"pcie_aspm_policy,omitempty"`
	EnergyPerfBias  string            `json:"energy_perf_bias,omitempty"`
	Mitigations     map[string]string `json:"mitigations,omitempty"`
	KernelCmdline   []string          `json:"kernel_cmdline,omitempty"`
}

// Power holds the throttling and energy counters.
type Power struct {
	ThrottleThermal uint64  `json:"throttle_thermal"`
	ThrottlePower   uint64  `json:"throttle_power"`
	RaplPackages    int     `json:"rapl_packages"`
	EnergyJoules    float64 `json:"rapl_energy_joules,omitempty"`
	RaplName        string  `json:"rapl_name,omitempty"`
}

// Virt describes the virtualization layer, if any.
type Virt struct {
	Hypervisor string `json:"hypervisor,omitempty"`
	Type       string `json:"type,omitempty"`
}
