package dsk

import (
	"encoding/binary"
	"fmt"

	"retroio/storage"
)

// The "Disc Information block" is always at offset 0 in the disk image file.
//
// * "MV - CPC" must be present, because it is used to identify the file as a disk image.
//   - "\r" is the C programming language equivalent of ASCII character 13.
//   - "\n" is the C programming language equivalent of ASCII character 10.
// * All tracks must be the same size.
// * "size of track" is used to calculate the location of the data for a chosen track.
// * "size of track" includes the &100 byte Track Information Block.
type DiskInformation struct {
	Identifier [34]byte  // "MV - CPCEMU Disk-File\r\nDisk-Info\r\n"
	Creator    [14]byte  // name of creator
	Tracks     uint8     // number of tracks
	Sides      uint8     // number of sides
	TrackSize  uint16    // size of a track
	Padding    [204]byte // unused padding, to Track Info block at 0x0100
}

// Read the disk information header.
func (d *DiskInformation) Read(reader *storage.Reader) error {
	return binary.Read(reader, binary.LittleEndian, d)
}

func (d DiskInformation) String() string {

	str := ""
	str += fmt.Sprintf("Identifier: %s\n", reformatIdentifier(d.Identifier[:]))
	str += fmt.Sprintf("Creator:    %s\n", d.Creator)
	str += fmt.Sprintf("Tracks:     %d\n", d.Tracks)
	str += fmt.Sprintf("Sides:      %d\n", d.Sides)
	str += fmt.Sprintf("Track Size: %d\n", d.TrackSize)
	return str
}
