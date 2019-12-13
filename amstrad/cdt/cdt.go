// Package cdt implements reading Amstrad CDT (TZX) formatted files,
// as specified in the TZX specification.
// https://www.worldofspectrum.org/TZXformat.html
//
// The `.CDT` tape image file format is identical to the `.TZX` file format designed by Tomaz Kac.
// Therefore this package is a simple wrapper around the `spectrum/tzx` package.
package cdt

import (
	"retroio/spectrum/tzx"
	"retroio/storage"
)

func New(reader *storage.Reader) *tzx.TZX {
	return tzx.New(reader)
}