package profile

// Profile is a declaration of how a machine is supposed to be built.
//
// It is the other half of asbuilt: the manifest says what the machine is,
// the profile says what it should be, and the diff between the two is the
// audit. Profiles are plain JSON so that a hardware team can keep them in
// version control next to the fleet inventory.
type Profile struct {
	Name   string `json:"name,omitempty"`
	Expect Expect `json:"expect"`
}

// Expect holds the expected values. Every field is optional: a field that
// is absent is not checked, and a field that is present is checked exactly.
type Expect struct {
	Platform *Platform `json:"platform,omitempty"`
	CPU      *CPU      `json:"cpu,omitempty"`
	Memory   *Memory   `json:"memory,omitempty"`
	PCIe     *PCIe     `json:"pcie,omitempty"`
	OS       *OS       `json:"os,omitempty"`
}

// Platform expects the firmware identity of the machine.
type Platform struct {
	Vendor      string `json:"vendor,omitempty"`
	Product     string `json:"product,omitempty"`
	BIOSVendor  string `json:"bios_vendor,omitempty"`
	BIOSVersion string `json:"bios_version,omitempty"`
}

// CPU expects the processor layout.
type CPU struct {
	Sockets        *int `json:"sockets,omitempty"`
	CoresPerSocket *int `json:"cores_per_socket,omitempty"`
	ThreadsPerCore *int `json:"threads_per_core,omitempty"`
	ModelContains  string `json:"model_contains,omitempty"`
	TurboEnabled   *bool `json:"turbo_enabled,omitempty"`
}

// Memory expects the memory population.
type Memory struct {
	Channels        *int   `json:"channels,omitempty"`
	DIMMsPerChannel *int   `json:"dimms_per_channel,omitempty"`
	SlotsFilled     *int   `json:"slots_filled,omitempty"`
	ConfiguredMTs   *int   `json:"configured_mts,omitempty"`
	TotalGB         *int   `json:"total_gb,omitempty"`
	ECCEnabled      *bool  `json:"ecc_enabled,omitempty"`
}

// PCIe expects the interconnect state.
type PCIe struct {
	MinLinkGTs  *float64 `json:"min_link_gts,omitempty"`
	MinLinkWidth *int     `json:"min_link_width,omitempty"`
	DegradedAllowed *bool `json:"degraded_allowed,omitempty"`
}

// OS expects the operating system configuration.
type OS struct {
	TransparentHugepage string `json:"transparent_hugepage,omitempty"`
	NumaBalancing       string `json:"numa_balancing,omitempty"`
	Governor            string `json:"governor,omitempty"`
}
