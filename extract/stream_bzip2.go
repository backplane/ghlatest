package extract

import (
	"compress/bzip2"
	"io/ioutil"
)

func (a *Archive) Bunzip2() error {
	// https://pkg.go.dev/compress/bzip2
	a.StreamHandle = ioutil.NopCloser(bzip2.NewReader(a.FileHandle))
	return nil
}
