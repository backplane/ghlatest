package extract

import (
	"compress/gzip"
)

func (a *Archive) Gunzip() error {
	// https://pkg.go.dev/compress/gzip@go1.20.1#example-package-WriterReader
	// Note the caller needs to close the reader
	zr, err := gzip.NewReader(a.FileHandle)
	if err != nil {
		return err
	}
	a.StreamHandle = zr
	return err
}
