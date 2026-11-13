package dmi

import (
	"testing"

	"github.com/Balanov/asbuilt/internal/dmitest"
)

// buildTable assembles a synthetic table that mirrors a two-socket server
// with eight DDR5 channels and two modules per channel.
func buildTable(t *testing.T) *Table {
	t.Helper()
	b := &dmitest.Builder{}
	b.BIOS("Dell Inc.", "2.10.1", "03/14/2024")
	b.System("Dell Inc.", "PowerEdge R760", "Not Specified", "CN7792", "4c4c4544", "0KJ3T2")
	b.Baseboard("Dell Inc.", "0KJ3T2", "A00")
	b.MemoryArray(0x1100, 1<<30, 16)
	for i, ch := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
		for slot := 1; slot <= 2; slot++ {
			b.MemoryDevice(dmitest.MemorySpec{
				Handle:        uint16(0x1200 + i*2 + slot),
				ArrayHandle:   0x1100,
				Locator:       "DIMM_" + ch + string(rune('0'+slot)),
				BankLocator:   "NODE 0 " + ch,
				SizeMB:        64 * 1024,
				MemoryType:    0x1B, // DDR5
				Vendor:        "Kingston",
				Serial:        "0A1B2C3D",
				PartNumber:    "KSM56R46BD4-256HA",
				RankCount:     2,
				SpeedMTs:      5600,
				ConfiguredMTs: 4800,
			})
		}
	}
	table, err := Parse(b.Build())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return table
}

func TestParseFindsTableInSMBIOS3Layout(t *testing.T) {
	table := buildTable(t)
	if table.SMBIOSVersion != 3 {
		t.Fatalf("SMBIOS version = %d, want 3", table.SMBIOSVersion)
	}
	if len(table.Entries) != 4+16 {
		t.Fatalf("entries = %d, want %d", len(table.Entries), 4+16)
	}
}

func TestParseFindsTableInSMBIOS2Layout(t *testing.T) {
	b := &dmitest.Builder{}
	b.BIOS("Dell Inc.", "2.10.1", "03/14/2024")
	b.System("Dell Inc.", "PowerEdge R760", "Not Specified", "CN7792", "4c4c4544", "0KJ3T2")
	table, err := Parse(b.BuildV2())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(table.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(table.Entries))
	}
}

func TestBIOSAndSystem(t *testing.T) {
	table := buildTable(t)

	bios, ok := table.BIOS()
	if !ok {
		t.Fatal("no BIOS structure")
	}
	if bios.Vendor != "Dell Inc." || bios.Version != "2.10.1" || bios.ReleaseDate != "03/14/2024" {
		t.Fatalf("BIOS = %+v", bios)
	}

	sys, ok := table.System()
	if !ok {
		t.Fatal("no system structure")
	}
	if sys.Product != "PowerEdge R760" || sys.Serial != "CN7792" || sys.SKU != "0KJ3T2" {
		t.Fatalf("system = %+v", sys)
	}
	if sys.Manufacturer != "Dell Inc." {
		t.Fatalf("manufacturer = %q", sys.Manufacturer)
	}
}

func TestMemoryDevicesCarryConfiguredAndRatedSpeed(t *testing.T) {
	table := buildTable(t)
	devices := table.MemoryDevices()
	if len(devices) != 16 {
		t.Fatalf("memory devices = %d, want 16", len(devices))
	}
	d := devices[0]
	if d.SizeMB != 64*1024 {
		t.Fatalf("size = %d MB, want %d", d.SizeMB, 64*1024)
	}
	if d.ConfiguredSpeedMTs != 4800 {
		t.Fatalf("configured speed = %d, want 4800", d.ConfiguredSpeedMTs)
	}
	if d.SpeedMTs != 5600 {
		t.Fatalf("rated speed = %d, want 5600", d.SpeedMTs)
	}
	if d.RankCount != 2 {
		t.Fatalf("ranks = %d, want 2", d.RankCount)
	}
	if d.MemoryType != "ddr5" {
		t.Fatalf("memory type = %q, want ddr5", d.MemoryType)
	}
	if d.Locator != "DIMM_A1" {
		t.Fatalf("locator = %q, want DIMM_A1", d.Locator)
	}
	if d.PartNumber != "KSM56R46BD4-256HA" {
		t.Fatalf("part number = %q", d.PartNumber)
	}
	if d.Vendor != "Kingston" {
		t.Fatalf("vendor = %q", d.Vendor)
	}
}

func TestMemoryArrayDeviceCount(t *testing.T) {
	table := buildTable(t)
	arrays := table.MemoryArrays()
	if len(arrays) != 1 {
		t.Fatalf("memory arrays = %d, want 1", len(arrays))
	}
	if arrays[0].DeviceCount != 16 {
		t.Fatalf("device count = %d, want 16", arrays[0].DeviceCount)
	}
	if arrays[0].Use != "system memory" {
		t.Fatalf("use = %q", arrays[0].Use)
	}
}

func TestChannelAndSocketOf(t *testing.T) {
	cases := []struct {
		locator string
		channel string
		socket  string
	}{
		{"DIMM_A1", "A", ""},
		{"DIMM_B2", "B", ""},
		{"CPU1_DIMM_A1", "A", "cpu1"},
		{"cpu2_dimm_h2", "H", "cpu2"},
		{"", "", ""},
	}
	for _, c := range cases {
		if got := ChannelOf(c.locator); got != c.channel {
			t.Errorf("ChannelOf(%q) = %q, want %q", c.locator, got, c.channel)
		}
		if got := SocketOf(c.locator); got != c.socket {
			t.Errorf("SocketOf(%q) = %q, want %q", c.locator, got, c.socket)
		}
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if _, err := Parse([]byte("not a dmi table at all")); err == nil {
		t.Fatal("expected an error for garbage input")
	}
}
