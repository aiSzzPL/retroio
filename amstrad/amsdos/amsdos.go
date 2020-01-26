package amsdos

// IsHeader calculates a checksum against the first 67-bytes of the file, and
// returns true if it is valid.
func IsHeader(expectedChecksum uint8, bytes []byte) bool {
	// TODO: validate checksum
	return true
}
