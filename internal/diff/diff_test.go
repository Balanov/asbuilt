package diff

import (
	"testing"

	"github.com/Balanov/asbuilt/internal/collect"
)

func manifest() *collect.Manifest {
	return &collect.Manifest{
		Schema: collect.SchemaVersion,
		Plat: collect.Platform{
			Vendor: "Dell Inc.", Product: "PowerEdge R760", Serial: "CN7792",
			BIOS: collect.BIOS{Vendor: "Dell Inc.", Version: "2.10.1", ReleaseDate: "03/14/2024"},
		},
		CPU: collect.CPU{
			Model: "Intel(R) Xeon(R) Platinum 8480+", Sockets: 2, CoresPerSocket: 28,
			ThreadsPerCore: 2, LogicalCPUs: 112, SMTEnabled: true,
			ISA: collect.ISA{Fingerprint: "abc123"},
		},
		NUMA: collect.NUMA{Nodes: 2},
		Memory: collect.Memory{
			TotalKB: 1030804372, SlotsFilled: 16, SlotsTotal: 16,
			ConfiguredMTs: 4800, MaxMTs: 5600, ECCEnabled: true,
			Channels: []collect.MemoryChannel{
				{ID: "A", DIMMs: 2, TotalMB: 131072, ConfiguredMTs: 4800},
			},
			DIMMs: []collect.MemoryDIMM{
				{Locator: "DIMM_A1", Channel: "A", SizeMB: 65536, PartNumber: "KSM56R46BD4-256HA"},
			},
		},
		PCIe: collect.PCIe{Devices: []collect.PCIeDevice{
			{Address: "0000:01:00.0", CurrentSpeedGTs: 32, CurrentWidth: 16, MaxSpeedGTs: 32, MaxWidth: 16},
		}},
		OS: collect.OS{Kernel: "Linux 6.1.0", THP: "always [madvise] never"},
	}
}

func TestIdenticalManifestsHaveNoDifferences(t *testing.T) {
	diffs := Manifests(manifest(), manifest())
	if len(diffs) != 0 {
		t.Fatalf("identical manifests produced %d differences: %+v", len(diffs), diffs)
	}
}

func TestDiffFindsTheChangedFields(t *testing.T) {
	a := manifest()
	b := manifest()
	b.Plat.BIOS.Version = "2.11.0"
	b.CPU.Sockets = 1
	b.Memory.ConfiguredMTs = 3600
	b.Memory.Channels[0].DIMMs = 1
	b.PCIe.Devices[0].CurrentSpeedGTs = 16

	diffs := Manifests(a, b)

	want := map[string]bool{
		"platform.bios.version":            false,
		"cpu.sockets":                      false,
		"memory.configured_mts":            false,
		"memory.channels.A.dimms":          false,
		"pcie.0000:01:00.0.current_speed_gts": false,
	}
	for _, d := range diffs {
		if _, ok := want[d.Path]; ok {
			want[d.Path] = true
		}
	}
	for path, found := range want {
		if !found {
			t.Errorf("the difference %s was not reported", path)
		}
	}

	for _, d := range diffs {
		if d.Path == "memory.configured_mts" && d.Impact != "memory bandwidth" {
			t.Errorf("the speed difference should name the bandwidth, got %q", d.Impact)
		}
	}
}

func TestDiffReportsModulesThatOnlyExistOnOneSide(t *testing.T) {
	a := manifest()
	b := manifest()
	b.Memory.DIMMs = append(b.Memory.DIMMs, collect.MemoryDIMM{
		Locator: "DIMM_B1", Channel: "B", SizeMB: 65536,
	})

	diffs := Manifests(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "memory.dimms.DIMM_B1" && d.Kind == "only_in_b" {
			found = true
		}
	}
	if !found {
		t.Errorf("the extra module was not reported: %+v", diffs)
	}
}
