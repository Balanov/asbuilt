// Package dmitest builds synthetic SMBIOS tables.
//
// The firmware table is the only source that describes the memory
// population of a server, and it is root-only on a live machine. Building
// one in memory lets the parser, the collectors and the rules be tested
// against exactly the bytes a server would publish.
//
// The field offsets follow the DMTF SMBIOS Reference Specification 3.9.0
// (DSP0134).
package dmitest

import (
	"encoding/binary"
)

// Entry is one SMBIOS structure.
type Entry struct {
	Type    byte
	Handle  uint16
	Data    []byte
	Strings []string
}

// Builder assembs an SMBIOS blob.
type Builder struct {
	entries []Entry
}

// Add appends a raw structure.
func (b *Builder) Add(typ byte, handle uint16, data []byte, strings ...string) *Builder {
	b.entries = append(b.entries, Entry{Type: typ, Handle: handle, Data: data, Strings: strings})
	return b
}

// setRefs writes one-based string references into the formatted area.
// The offset is the structure offset, as in the specification.
func setRefs(data []byte, offset int, first int, count int) {
	for i := 0; i < count; i++ {
		pos := offset - 4 + i
		if pos >= 0 && pos < len(data) {
			data[pos] = byte(first + i)
		}
	}
}

// BIOS appends a type 0 structure.
func (b *Builder) BIOS(vendor, version, releaseDate string) *Builder {
	data := make([]byte, 0x14)
	setRefs(data, 0x04, 1, 3)
	return b.Add(0, 0x0000, data, vendor, version, releaseDate)
}

// System appends a type 1 structure.
func (b *Builder) System(vendor, product, version, serial, uuid, sku string) *Builder {
	data := make([]byte, 0x17)
	setRefs(data, 0x04, 1, 6)
	return b.Add(1, 0x0100, data, vendor, product, version, serial, uuid, sku)
}

// Baseboard appends a type 2 structure.
func (b *Builder) Baseboard(vendor, product, version string) *Builder {
	data := make([]byte, 0x0B)
	setRefs(data, 0x04, 1, 3)
	return b.Add(2, 0x0200, data, vendor, product, version)
}

// MemoryArray appends a type 16 structure.
//
// Location, Use and Error Correction are enumeration values, not strings.
func (b *Builder) MemoryArray(handle uint16, maxCapacityKB int64, deviceCount int) *Builder {
	data := make([]byte, 0x15)
	data[0x00] = 0x03 // System board or motherboard
	data[0x01] = 0x03 // System memory
	data[0x02] = 0x06 // Multi-bit ECC
	binary.LittleEndian.PutUint32(data[0x03:0x07], uint32(maxCapacityKB))
	binary.LittleEndian.PutUint16(data[0x07:0x09], uint16(deviceCount))
	return b.Add(16, handle, data)
}

// MemorySpec describes a type 17 structure.
type MemorySpec struct {
	Handle          uint16
	ArrayHandle     uint16
	Locator         string
	BankLocator     string
	SizeMB          int
	MemoryType      byte
	Vendor          string
	Serial          string
	PartNumber      string
	RankCount       int
	SpeedMTs        int
	ConfiguredMTs   int
}

// MemoryDevice appends a type 17 structure.
func (b *Builder) MemoryDevice(s MemorySpec) *Builder {
	// A 3.3 style structure: long enough for the extended speed fields.
	data := make([]byte, 0x58)

	binary.LittleEndian.PutUint16(data[0x00:0x02], s.ArrayHandle)
	binary.LittleEndian.PutUint16(data[0x04:0x06], 72) // total width in bits
	binary.LittleEndian.PutUint16(data[0x06:0x08], 64) // data width in bits

	// Size at 0x0C, in megabytes, or the extended size for large modules.
	if s.SizeMB >= 0x7FFF {
		binary.LittleEndian.PutUint16(data[0x08:0x0A], 0x7FFF)
		binary.LittleEndian.PutUint32(data[0x18:0x1C], uint32(s.SizeMB))
	} else {
		binary.LittleEndian.PutUint16(data[0x08:0x0A], uint16(s.SizeMB))
	}

	data[0x0A] = 0x09 // Form factor: RDIMM
	data[0x0B] = 0x00 // Device set
	data[0x0E] = s.MemoryType
	binary.LittleEndian.PutUint16(data[0x11:0x13], uint16(s.SpeedMTs))

	// String references: locator, bank locator, manufacturer, serial,
	// asset tag, part number.
	setRefs(data, 0x10, 1, 2)
	setRefs(data, 0x17, 3, 4)

	// Attributes byte at structure offset 0x1B: bits 3:0 hold the rank count.
	data[0x1B-0x04] = byte(s.RankCount & 0x0F)

	// Configured memory speed at 0x20.
	binary.LittleEndian.PutUint16(data[0x1C:0x1E], uint16(s.ConfiguredMTs))

	return b.Add(17, s.Handle, data,
		s.Locator, s.BankLocator, s.Vendor, s.Serial, " ", s.PartNumber)
}

// Build assembles an SMBIOS 3.x blob: entry point followed by the table.
func (b *Builder) Build() []byte {
	return b.build(true)
}

// BuildV2 assembles an SMBIOS 2.x blob, which places the table at a
// different offset.
func (b *Builder) BuildV2() []byte {
	return b.build(false)
}

func (b *Builder) build(v3 bool) []byte {
	table := b.table()

	headerLen := 0x2E
	if !v3 {
		headerLen = 0x20
	}
	buf := make([]byte, headerLen)
	if v3 {
		copy(buf[0:], "_SM3_")
		buf[6] = 0x18 // entry point length
		buf[7] = 3    // major version
		buf[8] = 6    // minor version
		copy(buf[0x1A:], "_DMI_")
	} else {
		copy(buf[0:], "_SM_")
		buf[5] = 0x10 // entry point length
		buf[6] = 2    // major version
		buf[7] = 8    // minor version
		copy(buf[0x11:], "_DMI_")
	}
	return append(buf, table...)
}

// table serialises the structures and appends the end-of-table marker.
func (b *Builder) table() []byte {
	var out []byte
	for _, e := range b.entries {
		// The length field covers the whole structure: type, length, handle
		// and the formatted area.
		out = append(out, e.Type, byte(4+len(e.Data)))
		handle := make([]byte, 2)
		binary.LittleEndian.PutUint16(handle, e.Handle)
		out = append(out, handle...)
		out = append(out, e.Data...)
		for _, s := range e.Strings {
			out = append(out, s...)
			out = append(out, 0)
		}
		if len(e.Strings) == 0 {
			// A structure without strings still carries an empty string
			// table, which is two zero bytes.
			out = append(out, 0)
		}
		// One more zero closes the string table.
		out = append(out, 0)
	}
	// type 127, length 4, handle 0
	out = append(out, 127, 4, 0, 0)
	return out
}
