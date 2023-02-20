package extract

import (
	"compress/bzip2"
	"io/ioutil"
)

// Unbzip2 enables bzip2 decompression of the Archive data. It creates a bzip2
// stream reader for Archive.FileHandle on Archive.StreamHandle. Any errors
// creating the reader will be returned.
func (a *Archive) Unbzip2() error {
	// https://pkg.go.dev/compress/bzip2
	a.StreamHandle = ioutil.NopCloser(bzip2.NewReader(a.FileHandle))
	return nil
}
