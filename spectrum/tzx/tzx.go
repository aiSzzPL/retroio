// Package tzx implements reading of ZX Spectrum TZX formatted files,
// as specified in the TZX specification.
// https://www.worldofspectrum.org/TZXformat.html
//
// Rules and Definitions
//
//  * Any value requiring more than one byte is stored in little endian format (i.e. LSB first).
//  * Unused bits should be set to zero.
//  * Timings are given in Z80 clock ticks (T states) unless otherwise stated.
//      1 T state = (1/3500000)s
//  * Block IDs are given in hex.
//  * All ASCII texts use the ISO 8859-1 (Latin 1) encoding; some of them can have several lines, which
//    should be separated by ASCII code 13 decimal (0D hex).
//  * You might interpret 'full-period' as ----____ or ____----, and 'half-period' as ---- or ____.
//    One 'half-period' will also be referred to as a 'pulse'.
//  * Values in curly brackets {} are the default values that are used in the Spectrum ROM saving
//    routines. These values are in decimal.
//  * If there is no pause between two data blocks then the second one should follow immediately; not
//    even so much as one T state between them.
//  * This document refers to 'high' and 'low' pulse levels. Whether this is implemented as ear=1 and
//    ear=0 respectively or the other way around is not important, as long as it is done consistently.
//  * Zeros and ones in 'Direct recording' blocks mean low and high pulse levels respectively.
//    The 'current pulse level' after playing a Direct Recording block of CSW recording block
//    is the last level played.
//  * The 'current pulse level' after playing the blocks ID 10,11,12,13,14 or 19 is the opposite of
//    the last pulse level played, so that a subsequent pulse will produce an edge.
//  * A 'Pause' block consists of a 'low' pulse level of some duration. To ensure that the last edge
//    produced is properly finished there should be at least 1 ms. pause of the opposite level and only
//    after that the pulse should go to 'low'. At the end of a 'Pause' block the 'current pulse level'
//    is low (note that the first pulse will therefore not immediately produce an edge). A 'Pause' block
//    of zero duration is completely ignored, so the 'current pulse level' will NOT change in this case.
//    This also applies to 'Data' blocks that have some pause duration included in them.
//  * An emulator should put the 'current pulse level' to 'low' when starting to play a TZX file, either
//    from the start or from a certain position. The writer of a TZX file should ensure that the 'current
//    pulse level' is well-defined in every sequence of blocks where this is important, i.e. in any
//    sequence that includes a 'Direct recording' block, or that depends on edges generated by 'Pause'
//    blocks. The recommended way of doing this is to include a Pause after each sequence of blocks.
//  * When creating a 'Direct recording' block please stick to the standard sampling frequencies of 22050
//    or 44100 Hz. This will ensure correct playback when using PC's sound cards.
//  * The length of a block is given in the following format: numbers in square brackets [] mean that the
//    value must be read from the offset in the brackets. Other values are normal numbers.
//    Example: [02,03]+0A means: get number (a word) from offset 02 and add 0A. All numbers are in hex.
//  * General Extension Rule: ALL custom blocks that will be added after version 1.10 will have the length
//    of the block in first 4 bytes (long word) after the ID (this length does not include these 4 length
//    bytes). This should enable programs that can only handle older versions to skip that block.
//  * Just in case:
//      MSB = most significant byte
//      LSB = least significant byte
//      MSb = most significant bit
//      LSb = least significant bit
package tzx

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"

	"retroio/spectrum/basic"
	"retroio/spectrum/tap"
	"retroio/storage"
)

const (
	supportedMajorVersion = 1
	supportedMinorVersion = 20
)

// TZX files store the header information at the start of the file, followed
// by zero or more data blocks. Some TZX files include an ArchiveInfo block,
// which is always stored as the first block, directly after the header.
type TZX struct {
	reader *storage.Reader

	header
	archive Block
	blocks  []Block
}

// Block is an interface for Tape data blocks
type Block interface {
	Read(reader *storage.Reader)
	Id() uint8
	Name() string
	BlockData() tap.BlockI
}

// Header is the first block of data found in all TZX files.
// The file is identified with the first 7 bytes being `ZXTape!`, followed by the
// _end of file_ byte `26` (`1A` hex). This is followed by two bytes containing
// the major and minor version numbers of the TZX specification used.
type header struct {
	Signature    [7]byte // must be `ZXTape!`
	Terminator   uint8   // End of file marker
	MajorVersion uint8   // TZX major revision number
	MinorVersion uint8   // TZX minor revision number
}

func New(reader *storage.Reader) *TZX {
	return &TZX{reader: reader}
}

// Read processes the header, and then each block on the tape.
func (t *TZX) Read() error {
	if err := t.readHeader(); err != nil {
		return err
	}

	if err := t.readBlocks(); err != nil {
		return err
	}

	return nil
}

// readHeader reads the tape header data and validates that the format is correct.
func (t *TZX) readHeader() error {
	t.header = header{}

	if err := binary.Read(t.reader, binary.LittleEndian, &t.header); err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}

	if err := t.header.valid(); err != nil {
		return err
	}

	return nil
}

// readBlocks processes each TZX block on the tape.
func (t *TZX) readBlocks() error {
	for {
		_, err := t.reader.Peek(1)
		if err != nil && err == io.EOF {
			break // no problems, we're done!
		} else if err != nil {
			return err
		}
		blockID := t.reader.ReadByte()

		block, err := newFromBlockID(blockID)
		if err != nil {
			return err
		}
		block.Read(t.reader)

		if block.Id() == 0x32 {
			t.archive = block
		} else {
			t.blocks = append(t.blocks, block)
		}
	}
	return nil
}

// DisplayImageMetadata prints the metadata, archive info, data blocks, etc.
func (t TZX) DisplayImageMetadata() {
	if t.archive != nil {
		fmt.Println("ARCHIVE INFORMATION:")
		fmt.Println(t.archive)
	}

	fmt.Println("DATA BLOCKS:")
	for i, block := range t.blocks {
		fmt.Printf("#%02d %s\n", i+1, block)
	}

	fmt.Println()
	fmt.Printf("TZX revision: v%d.%d", t.MajorVersion, t.MinorVersion)
	if t.MinorVersion < supportedMinorVersion {
		fmt.Printf(
			" - WARNING! expected v%d.%d, this may lead to unexpected data or errors.",
			supportedMajorVersion,
			supportedMinorVersion,
		)
	}
	fmt.Println()
}

// ListBasicPrograms outputs all BASIC programs
func (t TZX) ListBasicPrograms() {
	isProgram := false
	filename := ""

	listing := ""
	for i, block := range t.blocks {
		if block.BlockData() == nil {
			continue
		}
		blk := block.BlockData()

		if isProgram == true {
			listing += fmt.Sprintf("BLK#%02d: %s\n", i+1, filename)

			program, err := basic.Decode(blk.BlockData())
			if err != nil {
				listing += fmt.Sprintf("    %s\n", err)
				continue
			}

			for _, line := range program {
				listing += line
			}
			listing += "\n"
			isProgram = false
		} else if blk.Id() == 0 && blk.Filename() != "" {
			filename = strings.Trim(block.BlockData().Filename(), " ")
			isProgram = true
		}
	}
	if len(listing) > 0 {
		fmt.Println("BASIC PROGRAMS:")
		fmt.Println()
		fmt.Println(listing)
	} else {
		fmt.Println("Unable to decode BASIC program")
	}
}

// Validates the TZX header data.
func (h header) valid() error {
	var validationError error

	sig := [7]byte{}
	copy(sig[:], "ZXTape!")
	if h.Signature != sig {
		validationError = errors.Wrapf(validationError, "Incorrect signature, got '%s'", h.Signature)
	}

	if h.Terminator != 0x1a {
		validationError = errors.Wrapf(validationError, "Incorrect terminator, got '%b'", h.Terminator)
	}

	if h.MajorVersion != supportedMajorVersion {
		validationError = errors.Wrapf(validationError, "Invalid version, got v%d.%d", h.MajorVersion, h.MinorVersion)
	}

	return validationError
}
