// Package dmi parses SMBIOS tables (the DMI data the firmware publishes).
//
// The firmware table is the only place a server tells you which DIMMs are
// actually populated, at which configured speed, and how many memory channels
// the board has. That is exactly the data a memory downclock audit needs, and
// no sysfs file exposes it.
//
// Parsing is done here instead of shelling out to dmidecode: one static
// binary, no fork, no dependency.
//
// The field offsets follow the DMTF SMBIOS Reference Specification 3.9.0
// (DSP0134), Table 75 for the memory device and Table 71 for the physical
// memory array.
package dmi

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// SMBIOS structure types used by this package.
const (
	TypeBIOS        = 0
	TypeSystem      = 1
	TypeBaseboard   = 2
	TypeProcessor   = 4
	TypeMemoryArray = 16
	TypeMemoryDevice = 17
)

// Entry is one SMBIOS structure.
type Entry struct {
	Type    byte
	Handle  uint16
	Length  byte
	Data    []byte
	Strings []string
}

// String returns the string at index i, or "" when it is absent.
func (e Entry) String(i int) string {
	if i < 0 || i >= len(e.Strings) {
		return ""
	}
	return e.Strings[i]
}

// Ref returns the string that the byte at the given structure offset refers
// to. SMBIOS string references are one-based indexes into the string table
// of the structure.
func (e Entry) Ref(offset int) string {
	if offset < 0 || offset >= len(e.Data) {
		return ""
	}
	idx := int(e.Data[offset])
	if idx == 0 || idx > len(e.Strings) {
		return ""
	}
	return e.Strings[idx-1]
}

// Table is a parsed SMBIOS table.
type Table struct {
	SMBIOSVersion byte
	Entries       []Entry
}

// Find returns every structure of the given type.
func (t *Table) Find(typ byte) []Entry {
	if t == nil {
		return nil
	}
	var out []Entry
	for _, e := range t.Entries {
		if e.Type == typ {
			out = append(out, e)
		}
	}
	return out
}

// Parse decodes an SMBIOS table as found in /sys/firmware/dmi/tables/DMI.
//
// The kernel exports the SMBIOS entry point together with the structure
// table, but the distance between them depends on the SMBIOS revision and
// on the firmware. Rather than hardcode an offset, the table is located by
// looking for the first offset where a structure stream parses cleanly and
// terminates with the end-of-table marker.
func Parse(b []byte) (*Table, error) {
	_, entries, ok := locateTable(b)
	if !ok {
		return nil, fmt.Errorf("dmi: no SMBIOS structure table found")
	}

	t := &Table{Entries: entries}
	switch {
	case len(b) >= 5 && string(b[0:5]) == "_SM3_":
		t.SMBIOSVersion = 3
	case len(b) >= 4 && string(b[0:4]) == "_SM_":
		t.SMBIOSVersion = 2
	}
	return t, nil
}

// locateTable finds the structure table inside the exported blob.
//
// A valid table is a run of structures that ends with the type 127
// end-of-table marker and starts with the type 0 BIOS structure, which the
// SMBIOS specification requires. Every candidate offset is tried and the
// richest valid table wins.
func locateTable(b []byte) (int, []Entry, bool) {
	candidates := make([]int, 0, 0x40)
	candidates = append(candidates, 0)
	for off := 0x08; off < 0x40; off++ {
		candidates = append(candidates, off)
	}
	bestOff, bestEntries := -1, []Entry(nil)
	for _, off := range candidates {
		if off >= len(b) {
			continue
		}
		entries, complete := parseStructures(b, off)
		if !complete || len(entries) == 0 {
			continue
		}
		if entries[0].Type != TypeBIOS {
			continue
		}
		if len(entries) > len(bestEntries) {
			bestOff, bestEntries = off, entries
		}
	}
	if bestOff < 0 {
		return 0, nil, false
	}
	return bestOff, bestEntries, true
}

// parseStructures reads structures from off. The second result reports
// whether the stream ended with the end-of-table marker.
func parseStructures(b []byte, off int) ([]Entry, bool) {
	var entries []Entry
	pos := off
	for pos+4 <= len(b) {
		typ := b[pos]
		length := b[pos+1]
		handle := binary.LittleEndian.Uint16(b[pos+2 : pos+4])
		if typ == 127 { // end of table
			return entries, true
		}
		if length < 4 || pos+int(length) > len(b) {
			return entries, false
		}
		data := b[pos : pos+int(length)]
		pos += int(length)

		// The string table is a list of null terminated strings followed by
		// one extra null byte. The terminator of the last string and that
		// extra byte are the two consecutive zeros that close the table.
		strStart := pos
		var strs []string
		for {
			if pos >= len(b) {
				return entries, false
			}
			if b[pos] == 0 {
				switch {
				case pos == strStart:
					// No strings at all: both zeros belong to the table.
					pos += 2
				case b[pos-1] == 0:
					pos++
				default:
					pos++
					continue
				}
				break
			}
			end := pos
			for end < len(b) && b[end] != 0 {
				end++
			}
			strs = append(strs, string(b[pos:end]))
			pos = end + 1
		}

		entries = append(entries, Entry{
			Type:    typ,
			Handle:  handle,
			Length:  length,
			Data:    data,
			Strings: strs,
		})
	}
	return entries, false
}

// BIOS is the type 0 structure.
type BIOS struct {
	Vendor      string
	Version     string
	ReleaseDate string
}

// System is the type 1 structure.
type System struct {
	Manufacturer string
	Product      string
	Version      string
	Serial       string
	UUID         string
	SKU          string
}

// Baseboard is the type 2 structure.
type Baseboard struct {
	Manufacturer string
	Product      string
	Version      string
}

// Processor is the type 4 structure.
type Processor struct {
	Socket       string
	Type         string
	Family       string
	Manufacturer string
	Version      string
	MaxSpeedMHz  int
	CoreCount    int
}

// MemoryDevice is the type 17 structure.
//
// SpeedMTs is the maximum speed the module itself supports.
// ConfiguredSpeedMTs is what the memory controller actually runs at today.
// When the controller downclocks, these two diverge, and that divergence is
// a measurable bandwidth loss rather than a matter of opinion.
type MemoryDevice struct {
	Locator            string
	BankLocator        string
	SizeMB             int
	FormFactor         string
	MemoryType         string
	TypeDetail         string
	Vendor             string
	Serial             string
	AssetTag           string
	PartNumber         string
	RankCount          int
	SpeedMTs           int
	ConfiguredSpeedMTs int
	ArrayHandle        uint16
}

// MemoryArray is the type 16 structure.
type MemoryArray struct {
	Handle          uint16
	Location        string
	Use             string
	ErrorCorrection string
	MaxCapacityKB   int64
	DeviceCount     int
}

// BIOS decodes the type 0 structure.
func (t *Table) BIOS() (BIOS, bool) {
	entries := t.Find(TypeBIOS)
	if len(entries) == 0 {
		return BIOS{}, false
	}
	e := entries[0]
	return BIOS{
		Vendor:      e.Ref(0x04),
		Version:     e.Ref(0x05),
		ReleaseDate: e.Ref(0x06),
	}, true
}

// System decodes the type 1 structure.
func (t *Table) System() (System, bool) {
	entries := t.Find(TypeSystem)
	if len(entries) == 0 {
		return System{}, false
	}
	e := entries[0]
	return System{
		Manufacturer: e.Ref(0x04),
		Product:      e.Ref(0x05),
		Version:      e.Ref(0x06),
		Serial:       e.Ref(0x07),
		UUID:         e.Ref(0x08),
		SKU:          e.Ref(0x09),
	}, true
}

// Baseboard decodes the type 2 structure.
func (t *Table) Baseboard() (Baseboard, bool) {
	entries := t.Find(TypeBaseboard)
	if len(entries) == 0 {
		return Baseboard{}, false
	}
	e := entries[0]
	return Baseboard{
		Manufacturer: e.Ref(0x04),
		Product:      e.Ref(0x05),
		Version:      e.Ref(0x06),
	}, true
}

// Processors decodes every type 4 structure.
func (t *Table) Processors() []Processor {
	var out []Processor
	for _, e := range t.Find(TypeProcessor) {
		p := Processor{
			Socket:       e.Ref(0x04),
			Type:         e.Ref(0x05),
			Family:       e.Ref(0x06),
			Manufacturer: e.Ref(0x07),
			Version:      e.Ref(0x08),
		}
		if len(e.Data) >= 0x18 {
			p.MaxSpeedMHz = int(binary.LittleEndian.Uint16(e.Data[0x16:0x18]))
		}
		out = append(out, p)
	}
	return out
}

// MemoryDevices decodes every populated type 17 structure.
func (t *Table) MemoryDevices() []MemoryDevice {
	var out []MemoryDevice
	for _, e := range t.Find(TypeMemoryDevice) {
		d := MemoryDevice{
			Locator:     e.Ref(0x10),
			BankLocator: e.Ref(0x11),
			Vendor:      e.Ref(0x17),
			Serial:      e.Ref(0x18),
			AssetTag:    e.Ref(0x19),
			PartNumber:  e.Ref(0x1A),
		}
		if len(e.Data) >= 0x06 {
			d.ArrayHandle = binary.LittleEndian.Uint16(e.Data[0x04:0x06])
		}
		d.SizeMB = memoryDeviceSizeMB(e)
		if len(e.Data) >= 0x0F {
			d.FormFactor = formFactor(e.Data[0x0E])
			d.MemoryType = memoryType(e.Data[0x12])
		}
		if len(e.Data) >= 0x17 {
			d.SpeedMTs = int(binary.LittleEndian.Uint16(e.Data[0x15:0x17]))
		}
		if len(e.Data) >= 0x1C {
			// Attributes byte, bits 3:0 hold the rank count.
			d.RankCount = int(e.Data[0x1B] & 0x0F)
		}
		if len(e.Data) >= 0x22 {
			d.ConfiguredSpeedMTs = int(binary.LittleEndian.Uint16(e.Data[0x20:0x22]))
		}
		if d.SizeMB <= 0 {
			continue
		}
		out = append(out, d)
	}
	return out
}

// memoryDeviceSizeMB decodes the size of a memory device.
//
// The Size field holds megabytes, or kilobytes when bit 15 is set. Devices
// of 32 GiB and above report the 0x7FFF marker and keep the real size in the
// Extended Size field, which counts megabytes.
func memoryDeviceSizeMB(e Entry) int {
	if len(e.Data) < 0x0E {
		return 0
	}
	raw := binary.LittleEndian.Uint16(e.Data[0x0C:0x0E])
	switch {
	case raw == 0:
		return 0
	case raw == 0xFFFF:
		return 0
	case raw == 0x7FFF:
		if len(e.Data) < 0x20 {
			return 0
		}
		ext := binary.LittleEndian.Uint32(e.Data[0x1C:0x20]) & 0x7FFFFFFF
		return int(ext)
	case raw&0x8000 != 0:
		return int(raw&0x7FFF) / 1024
	default:
		return int(raw)
	}
}

// MemoryArrays decodes every type 16 structure.
func (t *Table) MemoryArrays() []MemoryArray {
	var out []MemoryArray
	for _, e := range t.Find(TypeMemoryArray) {
		a := MemoryArray{Handle: e.Handle}
		if len(e.Data) >= 0x07 {
			a.Location = arrayLocation(e.Data[0x04])
			a.Use = arrayUse(e.Data[0x05])
			a.ErrorCorrection = arrayErrorCorrection(e.Data[0x06])
		}
		if len(e.Data) >= 0x0B {
			a.MaxCapacityKB = int64(binary.LittleEndian.Uint32(e.Data[0x07:0x0B]))
		}
		if len(e.Data) >= 0x0D {
			a.DeviceCount = int(binary.LittleEndian.Uint16(e.Data[0x0B:0x0D]))
		}
		out = append(out, a)
	}
	return out
}

func arrayLocation(b byte) string {
	switch b {
	case 0x03:
		return "system board"
	case 0x04:
		return "isa add-on card"
	case 0x05:
		return "eisa add-on card"
	case 0x06:
		return "pci add-on card"
	case 0x07:
		return "mca add-on card"
	case 0x08:
		return "pcmcia add-on card"
	case 0x09:
		return "proprietary add-on card"
	case 0xA4:
		return "cxl add-on card"
	}
	return fmt.Sprintf("0x%02x", b)
}

func arrayUse(b byte) string {
	switch b {
	case 0x03:
		return "system memory"
	case 0x04:
		return "video memory"
	case 0x05:
		return "flash memory"
	case 0x06:
		return "non-volatile ram"
	case 0x07:
		return "cache memory"
	}
	return fmt.Sprintf("0x%02x", b)
}

func arrayErrorCorrection(b byte) string {
	switch b {
	case 0x03:
		return "none"
	case 0x04:
		return "parity"
	case 0x05:
		return "single-bit ecc"
	case 0x06:
		return "multi-bit ecc"
	case 0x07:
		return "crc"
	}
	return fmt.Sprintf("0x%02x", b)
}

func formFactor(b byte) string {
	switch b {
	case 0x01:
		return "other"
	case 0x02:
		return "unknown"
	case 0x09:
		return "rdimm"
	case 0x0A:
		return "udimm"
	case 0x0B:
		return "so-dimm"
	case 0x0C:
		return "micro-dimm"
	case 0x0D:
		return "lrdimm"
	case 0x0E:
		return "cldimm"
	}
	return fmt.Sprintf("0x%02x", b)
}

func memoryType(b byte) string {
	switch b {
	case 0x1A:
		return "ddr4"
	case 0x1B:
		return "ddr5"
	case 0x1C:
		return "ddr3"
	case 0x1D:
		return "ddr2"
	case 0x1E:
		return "ddr"
	case 0x1F:
		return "ddr-2"
	case 0x20:
		return "zram"
	}
	return fmt.Sprintf("0x%02x", b)
}

// ChannelOf derives the memory channel of a DIMM from its slot locator.
//
// Board vendors name slots consistently: "DIMM_A1" is channel A, slot 1.
// "CPU1_DIMM_B2" is channel B of the first socket. The trailing digits are
// the slot within the channel.
func ChannelOf(locator string) string {
	s := strings.ToUpper(strings.TrimSpace(locator))
	if s == "" {
		return ""
	}
	end := len(s)
	for end > 0 && s[end-1] >= '0' && s[end-1] <= '9' {
		end--
	}
	name := strings.Trim(s[:end], "_- ")
	if i := strings.LastIndex(name, "_"); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// SocketOf derives the socket of a DIMM from its slot locator when the
// locator names one ("CPU1_DIMM_A1", "DIMM_A1" for a single socket).
func SocketOf(locator string) string {
	s := strings.ToUpper(strings.TrimSpace(locator))
	if strings.HasPrefix(s, "CPU") {
		rest := s[3:]
		end := 0
		for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
			end++
		}
		if end > 0 {
			return "cpu" + rest[:end]
		}
	}
	return ""
}
