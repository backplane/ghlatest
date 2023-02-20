package extract

import (
	"io/ioutil"

	"github.com/ulikunitz/xz"
)

// Unxz enables xz decompression of the Archive data. It creates an xz stream
// reader for Archive.FileHandle on Archive.StreamHandle. Any errors creating
// the reader will be returned.
func (a *Archive) Unxz() error {
	// https://pkg.go.dev/github.com/ulikunitz/xz#section-readme
	xzr, err := xz.NewReader(a.FileHandle)
	if err != nil {
		return err
	}
	a.StreamHandle = ioutil.NopCloser(xzr)
	return err
}
