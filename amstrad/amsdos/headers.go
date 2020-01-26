// Amstrad CP/M Disc Format
//
// Reference: http://www.seasip.info/Cpm/amsform.html
package amsdos

import "retroio/cpm/cpm3"

// Amstrad CP/M (and +3DOS) has an eXtended Disc Parameter Block (XDPB):
//
// NOTE: The DPB is not stored on disc.
//
// CPC format detection:
// This simple system is used by CPC computers if the first physical sector is:
// - 41h: A System formatted disc:
//          single sided, single track, 40 tracks, 9 sectors/track, 512-byte sectors,
//          2 reserved tracks, 1k blocks,
//          2 directory blocks,
//          gap lengths 2Ah and 52h,
//          bootable
// - C1h: A Data formatted disc:
//          single sided, single track, 40 tracks, 9 sectors/track, 512-byte sectors,
//          no reserved tracks, 1k blocks,
//          2 directory blocks,
//          gap lengths 2Ah and 52h,
//          not bootable
//
// PCW/Spectrum format detection: see `PcwSpectrumRecord` below.
type DiscParameterBlock struct {
	cpm3.DiskParameterBlock

	// Amstrad eXtended parameters

	// Type of disc media (sidedness)
	//
	// Bit | Description
	// 0-1   0 => Single sided
	//       1 => Double sided, flip sides
	//          ie track   0 is cylinder   0 head 0
	//             track   1 is cylinder   0 head 1
	//             track   2 is cylinder   1 head 0
	//             ...
	//             track n-1 is cylinder n/2 head 0
	//             track   n is cylinder n/2 head 1
	//       2 => Double sided, up and over
	//          ie track   0 is cylinder 0 head 0
	//             track   1 is cylinder 1 head 0
	//             track   2 is cylinder 2 head 0
	//             ...
	//             track n-2 is cylinder 2 head 1
	//             track n-1 is cylinder 1 head 1
	//             track   n is cylinder 0 head 1
	//  6    Set if the format is for a high-density disc
	//         This is an extension in PCW16 CP/M, BIOS 0.09+.
	//         It is not an official part of the spec.
	//  7    Set if the format is double track.
	MediaType uint8

	// tracks/side
	TrackCountPerSide uint8

	// sectors/track
	SectorCountPerTrack uint8

	// first physical sector number
	FirstSectorNumber uint8

	// sector size, bytes
	SectorSize uint16

	// uPD765A read/write gap
	ReadWriteGap uint8

	// uPD765A format gap
	FormatGap uint8

	// MFM/Multitrack flags byte
	// Bit 7 set => Multitrack else Single track
	//     6 set => MFM mode else FM mode
	//     5 set => Skip deleted data address mark
	MultiTrackFlags uint8

	// freeze flag
	// Set to non-zero value to force this format to be used - otherwise,
	// attempt to determine format when a disc is logged in.
	FreezeFlag uint8
}

// AMDSDOS File Record Header
//
// Files may, or may not, have a header depending on the contents of the
// file - CP/M files do not have headers. This will not cause problems for
// programs written in BASIC but it is an important difference between
// cassette and disc files.
//
// AMSDOS files have a single header in the first 128 bytes of the file - the
// header record - except unprotected ASCII files, which have no header.
//
// These headers are detected by calculating the checksum the first 67 bytes of
// the record. If the checksum is as expected then a header is present, if not
// then there is no header. Thus it is possible, though unlikely, that a file
// without a header could be mistaken for one with a header.
type RecordHeader struct {
	// Cassette/Disc header
	User          uint8     // User number, #00..#0F
	Name          [8]uint8  // Name part, padded with spaces
	Type          [3]uint8  // Type part, padded with spaces
	Unknown       [4]uint8  // #00
	BlockNumber   uint8     // Not used, set to 0
	LastBlock     uint8     // Not used, set to 0
	FileType      uint8     // As per cassette
	DataLength    uint16    // As per cassette
	DataLocation  uint16    // As per cassette
	FirstBlock    uint8     // Set to #FF, only used for output files
	LogicalLength uint16    // As per cassette
	EntryAddress  uint16    // As per cassette
	Unallocated   [36]uint8 // As per cassette

	FileLength [3]uint8  // 24-bit value. Length of the file in bytes, excluding the header record. Least significant byte in lowest address.
	Checksum   uint16    // Sixteen bit checksum, sum of bytes 0..66
	Undefined  [58]uint8 // 69... 127 Undefined
}

// When a file without a header is opened for input a fake header is constructed in store.
// TODO: probably not needed, just use the normal disc header
type HeaderlessHeader struct {
	// Filename
	User    uint8    // User number, #00..#0F
	Name    [8]uint8 // Name part, padded with spaces
	Type    [3]uint8 // Type part, padded with spaces
	Unknown [4]uint8 // #00

	Unused1      uint8 // Not used, set to 0
	Unused2      uint8 // Not used, set to 0
	FileType     uint8 // #16, unprotected ASCII version 1
	Unused3      uint16
	DataLocation uint16 // Address of 2K buffer
	FirstBlock   uint8  // #FF
	Unused4      uint16
	Unused5      uint16
	Unused6      [36]uint8
}

// PCW/Spectrum system
//
// In addition to the XDPB system, the PCW and Spectrum +3 can determine the format
// of a disc from a 16-byte record on track 0, head 0, physical sector 1.
//
// If all bytes of the spec are 0E5h, it should be assumed that the disc is a
// 173k PCW/Spectrum +3 disc, ie:
//   single sided, single track, 40 tracks, 9 sectors/track, 512-byte sectors,
//   1 reserved track, 1k blocks,
//   2 directory blocks,
//   gap lengths 2Ah and 52h,
//   not bootable
//
// PCW16 extended boot record
//
// The "boot record" system has been extended in PCW16 CP/M (BIOS 0.09 and later).
// The extension is intended to allow a CP/M "partition" on a DOS-formatted floppy disc.
//
// An extended boot sector (cylinder 0, head 0, sector 1) has the following characteristics:
// - First byte is 0E9h or 0EBh
// - Where DOS expects the disc label to be (at sector + 2Bh) there are 11 ASCII bytes
//   of the form `CP/M????DSK`, where "?" can be any character.
// - At sector + 7Ch are the four ASCII bytes "CP/M"
// - At sector + 80h is the disc specification as described above.
type PcwSpectrumDPB struct {
	// format number
	//   0: SS SD
	//   1: CPC formats, but those formats don't have boot records anyway.
	//   2: ^
	//   3: DS DD
	// Any other value: bad format
	FormatNumber uint8

	// sidedness ; As in XDPB
	MediaType uint8

	// tracks/side
	TrackCountPerSide uint8

	// sectors/track
	SectorCountPerTrack uint8

	// physical sector shift ; psh in XDPB
	PhysicalShift uint8

	// no. reserved tracks ; off in XDPB
	ReservedTracks uint8

	// block shift ; bsh in XDPB
	BlockShift uint8

	// no. directory blocks
	DirectoryBlockCount uint8

	// uPD765A read/write gap length
	ReadWriteGap uint8

	// uPD765A format gap length
	FormatGap uint8

	// 0,0,0,0,0 ; Unused
	Unused [5]uint8

	// Checksum fiddle byte ; Used to indicate Bootable discs.
	//
	// Change this byte so that the 8-bit checksum of the sector is:
	//    1 - sector contains a PCW9512 bootstrap
	//    3 - sector contains a Spectrum +3 bootstrap
	//  255 - sector contains a PCW8256 bootstrap
	//        (the bootstrap code is in the remainder of the sector)
	Checksum uint8
}
