package rules

import (
	"math"
	"testing"

	"github.com/Balanov/asbuilt/internal/collect"
	"github.com/Balanov/asbuilt/internal/profile"
)

// baseManifest returns a machine with nothing wrong with it.
func baseManifest() *collect.Manifest {
	return &collect.Manifest{
		Schema: collect.SchemaVersion,
		CPU: collect.CPU{
			Model: "Intel(R) Xeon(R) Platinum 8480+",
			Sockets: 2, CoresPerSocket: 28, ThreadsPerCore: 2, LogicalCPUs: 112,
			SMTEnabled: true, BaseFreqMHz: 2000, MaxFreqMHz: 3500, TurboEnabled: true,
			Governor: "performance",
			ISA:      collect.ISA{Exposed: []string{"avx2", "sse4_2"}, Count: 2, Fingerprint: "abc"},
		},
		NUMA: collect.NUMA{Nodes: 2, Topology: []collect.NUMANode{
			{ID: 0, MemTotalKB: 515402186},
			{ID: 1, MemTotalKB: 515402186},
		}},
		Memory: collect.Memory{
			TotalKB: 1030804372, Source: "smbios", SlotsTotal: 16, SlotsFilled: 16,
			ConfiguredMTs: 4800, MaxMTs: 5600, ECCEnabled: true,
			BandwidthGBs: 307.2, BandwidthAtMaxGBs: 358.4,
			Channels: []collect.MemoryChannel{
				{ID: "A", DIMMs: 2, TotalMB: 131072, ConfiguredMTs: 4800, MaxMTs: 5600, PartNumbers: []string{"KSM56R46BD4-256HA"}},
				{ID: "B", DIMMs: 2, TotalMB: 131072, ConfiguredMTs: 4800, MaxMTs: 5600, PartNumbers: []string{"KSM56R46BD4-256HA"}},
			},
			DIMMs: []collect.MemoryDIMM{
				{Locator: "DIMM_A1", Channel: "A", SizeMB: 65536, ConfiguredSpeedMTs: 4800, MaxSpeedMTs: 5600, PartNumber: "KSM56R46BD4-256HA"},
				{Locator: "DIMM_A2", Channel: "A", SizeMB: 65536, ConfiguredSpeedMTs: 4800, MaxSpeedMTs: 5600, PartNumber: "KSM56R46BD4-256HA"},
				{Locator: "DIMM_B1", Channel: "B", SizeMB: 65536, ConfiguredSpeedMTs: 4800, MaxSpeedMTs: 5600, PartNumber: "KSM56R46BD4-256HA"},
				{Locator: "DIMM_B2", Channel: "B", SizeMB: 65536, ConfiguredSpeedMTs: 4800, MaxSpeedMTs: 5600, PartNumber: "KSM56R46BD4-256HA"},
			},
		},
		PCIe: collect.PCIe{Devices: []collect.PCIeDevice{
			{Address: "0000:01:00.0", CurrentSpeedGTs: 32, CurrentWidth: 16, MaxSpeedGTs: 32, MaxWidth: 16, PortMaxSpeedGTs: 32, PortMaxWidth: 16},
		}},
		OS: collect.OS{Kernel: "Linux 6.1.0", THP: "always [madvise] never", NUMABalancing: "0", ASPM: "performance", EnergyPerfBias: "performance"},
	}
}

func run(t *testing.T, m *collect.Manifest, p *profile.Profile) []Finding {
	t.Helper()
	return Run(&Context{Manifest: m, Profile: p})
}

func findByID(findings []Finding, id string) *Finding {
	for i := range findings {
		if findings[i].ID == id {
			return &findings[i]
		}
	}
	return nil
}

func TestCleanMachineHasNoCriticalFindings(t *testing.T) {
	findings := run(t, baseManifest(), nil)
	for _, f := range findings {
		if f.Severity == Critical {
			t.Errorf("unexpected critical finding %s: %s", f.ID, f.Detail)
		}
	}
}

func TestCatalogHasUniqueSortedIDs(t *testing.T) {
	catalog := Catalog()
	if len(catalog) < 20 {
		t.Fatalf("catalog has %d rules, want at least 20", len(catalog))
	}
	seen := map[string]bool{}
	prev := ""
	for _, r := range catalog {
		if r.ID == "" {
			t.Fatal("a rule has no id")
		}
		if seen[r.ID] {
			t.Fatalf("duplicate rule id %s", r.ID)
		}
		seen[r.ID] = true
		if r.ID < prev {
			t.Fatalf("catalog is not sorted: %s after %s", r.ID, prev)
		}
		prev = r.ID
		if r.Severity == "" {
			t.Errorf("rule %s has no severity", r.ID)
		}
		if r.Source == "" {
			t.Errorf("rule %s has no source", r.ID)
		}
		if r.Check == nil {
			t.Errorf("rule %s has no check", r.ID)
		}
	}
}

func TestPCIELinkDegradedQuantifiesTheLoss(t *testing.T) {
	m := baseManifest()
	m.PCIe.Devices[0].CurrentSpeedGTs = 16 // negotiated Gen4 instead of Gen5

	f := findByID(run(t, m, nil), "pcie.link.degraded")
	if f == nil {
		t.Fatal("the degraded link was not reported")
	}
	if f.Severity != Critical {
		t.Errorf("severity = %s, want critical", f.Severity)
	}
	if f.Impact == nil {
		t.Fatal("the finding carries no impact")
	}
	if f.Impact.Kind != "pcie_bandwidth" {
		t.Errorf("impact kind = %s, want pcie_bandwidth", f.Impact.Kind)
	}
	// Gen5 x16 is 63 GB/s per direction, Gen4 x16 is 31.5 GB/s.
	if math.Abs(f.Impact.LossPct-50) > 0.5 {
		t.Errorf("loss = %.1f%%, want about 50%%", f.Impact.LossPct)
	}
	if f.Impact.Potential < 62 || f.Impact.Actual < 31 {
		t.Errorf("bandwidth numbers look wrong: %.1f of %.1f GB/s", f.Impact.Actual, f.Impact.Potential)
	}
}

func TestPCIEPortCapIsReported(t *testing.T) {
	m := baseManifest()
	m.PCIe.Devices[0].PortMaxSpeedGTs = 16 // the board pins the slot to Gen4

	f := findByID(run(t, m, nil), "pcie.link.port_capped")
	if f == nil {
		t.Fatal("the capped port was not reported")
	}
	if f.Impact == nil || f.Impact.LossPct < 49 {
		t.Errorf("the port cap should cost about half the bandwidth, got %+v", f.Impact)
	}
}

func TestMemoryBelowModuleRatingQuantifiesTheLoss(t *testing.T) {
	m := baseManifest()
	findings := run(t, m, nil)
	f := findByID(findings, "mem.speed.below_module_rating")
	if f == nil {
		t.Fatal("the downclock was not reported")
	}
	if f.Impact == nil {
		t.Fatal("the finding carries no impact")
	}
	// 4800 of 5600 MT/s is 14.3% of the bandwidth gone.
	if math.Abs(f.Impact.LossPct-14.3) > 0.5 {
		t.Errorf("loss = %.1f%%, want about 14.3%%", f.Impact.LossPct)
	}
}

func TestUnevenPopulationIsCritical(t *testing.T) {
	m := baseManifest()
	m.Memory.Channels[1].DIMMs = 1
	m.Memory.Channels[1].TotalMB = 65536

	f := findByID(run(t, m, nil), "mem.population.uneven")
	if f == nil {
		t.Fatal("the uneven population was not reported")
	}
	if f.Severity != Critical {
		t.Errorf("severity = %s, want critical", f.Severity)
	}
	if f.Impact == nil || f.Impact.LossPct < 49 {
		t.Errorf("half the channels are empty, so about half the bandwidth is gone, got %+v", f.Impact)
	}
}

func TestSingleNUMANodeOnTwoSocketsIsCritical(t *testing.T) {
	m := baseManifest()
	m.NUMA.Nodes = 1
	m.NUMA.Topology = m.NUMA.Topology[:1]
	m.NUMA.Interleaved = true

	f := findByID(run(t, m, nil), "numa.single_node_multi_socket")
	if f == nil {
		t.Fatal("the hidden topology was not reported")
	}
	if f.Severity != Critical {
		t.Errorf("severity = %s, want critical", f.Severity)
	}
}

func TestThrottlingCountersAreCritical(t *testing.T) {
	m := baseManifest()
	m.Power.ThrottleThermal = 3
	m.Power.ThrottlePower = 17

	findings := run(t, m, nil)
	if findByID(findings, "power.throttle.thermal") == nil {
		t.Error("thermal throttling was not reported")
	}
	f := findByID(findings, "power.throttle.limit")
	if f == nil {
		t.Fatal("power limit throttling was not reported")
	}
	if f.Severity != Critical {
		t.Errorf("severity = %s, want critical", f.Severity)
	}
}

func TestProfileConformance(t *testing.T) {
	m := baseManifest()
	m.CPU.Sockets = 1
	m.Memory.ConfiguredMTs = 3600

	channels := 8
	dpc := 2
	mts := 4800
	wantSockets := 2
	p := &profile.Profile{
		Name: "xd-node",
		Expect: profile.Expect{
			CPU:    &profile.CPU{Sockets: &wantSockets},
			Memory: &profile.Memory{Channels: &channels, DIMMsPerChannel: &dpc, ConfiguredMTs: &mts},
		},
	}

	findings := run(t, m, p)
	if findByID(findings, "profile.cpu.sockets") == nil {
		t.Error("the socket mismatch was not reported")
	}
	if findByID(findings, "profile.memory.channels") == nil {
		t.Error("the channel mismatch was not reported")
	}
	f := findByID(findings, "profile.memory.configured_mts")
	if f == nil {
		t.Fatal("the speed mismatch was not reported")
	}
	if f.Impact == nil {
		t.Fatal("the speed mismatch carries no impact")
	}
	// 3600 of 4800 MT/s is 25% of the bandwidth gone.
	if math.Abs(f.Impact.LossPct-25) > 0.5 {
		t.Errorf("loss = %.1f%%, want 25%%", f.Impact.LossPct)
	}
}

func TestProfileRulesStaySilentWithoutAProfile(t *testing.T) {
	findings := run(t, baseManifest(), nil)
	for _, f := range findings {
		if len(f.ID) >= 7 && f.ID[:7] == "profile" {
			t.Errorf("profile rule %s fired without a profile", f.ID)
		}
	}
}

func TestFindingsAreSortedBySeverity(t *testing.T) {
	m := baseManifest()
	m.PCIe.Devices[0].CurrentSpeedGTs = 16
	m.Power.ThrottlePower = 5
	m.OS.THP = "always [always] never"

	findings := run(t, m, nil)
	prev := 4
	for _, f := range findings {
		if f.Severity.Rank() > prev {
			t.Fatalf("findings are not sorted by severity: %s after rank %d", f.ID, prev)
		}
		prev = f.Severity.Rank()
	}
}
