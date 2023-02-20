package extract

import (
	"compress/gzip"
)

// Ungzip enables gzip decompression of the Archive data. It creates a gzip stream
// reader for Archive.FileHandle on Archive.StreamHandle. Any errors creating
// the reader will be returned.
func (a *Archive) Ungzip() error {
	// https://pkg.go.dev/compress/gzip@go1.20.1#example-package-WriterReader
	// Note the caller needs to close the reader
	zr, err := gzip.NewReader(a.FileHandle)
	if err != nil {
		return err
	}
	a.StreamHandle = zr
	return err
}
