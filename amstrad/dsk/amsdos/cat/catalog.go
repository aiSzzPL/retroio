package cat

import (
	"errors"
	"retroio/amstrad/dsk/amsdos"
	"sort"
)

type catalog struct {
	Drive     byte
	User      uint8
	FreeSpace uint16
	Records   []directoryRecord
}

// directoryRecord is the displayable data for a directory record.
// This is similar to the CP/M Directory, except each entry merges all record extents.
type directoryRecord struct {
	Filename    [8]uint8
	FileType    [3]uint8
	RecordCount uint16 // Total record count for all extents of a record
}

// COMMAND: CAT
// Catalogs the disc. Generates a list, in alpha-numeric order, the full names
// of all files found, together with each file's length (to the nearest higher Kbyte).
// The free space left on the disc is also displayed, together with Drive and
// User identification.
func CommandCat(diskMaxBlocks uint16, directories []amsdos.Directory) (*catalog, error) {
	if len(directories) == 0 {
		return nil, errors.New("no directories found")
	}

	cat := &catalog{
		Drive:     'A',
		User:      directories[0].UserNumber,
		FreeSpace: diskMaxBlocks,
	}

	wasExtent := false
	var lastFilename [8]byte
	var lastFileType [3]byte

	for _, d := range directories {
		record := directoryRecord{
			Filename:    d.Filename,
			FileType:    d.FileType,
			RecordCount: cat.blockCount(d.Allocation),
		}
		cat.FreeSpace -= record.RecordCount

		if lastFilename == record.Filename && lastFileType == record.FileType {
			cat.Records[len(cat.Records)-1].RecordCount += record.RecordCount
			wasExtent = true
		} else {
			cat.Records = append(cat.Records, record)
		}

		lastFilename = record.Filename
		lastFileType = record.FileType
	}

	if wasExtent {
		cat.FreeSpace -= 1
	}

	cat.alphabetize()

	return cat, nil
}

func (c catalog) blockCount(allocation [16]uint8) uint16 {
	var blocks uint16
	for _, b := range allocation {
		if b > 0 {
			blocks += 1
		}
	}
	return blocks
}

func (c *catalog) alphabetize() {
	sort.Slice(c.Records, func(i, j int) bool {
		if c.Records[i].Filename == c.Records[j].Filename {
			return string(c.Records[i].FileType[:]) < string(c.Records[j].FileType[:])
		}
		return string(c.Records[i].Filename[:]) < string(c.Records[j].Filename[:])
	})
}
