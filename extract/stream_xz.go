package extract

import (
	"io/ioutil"

	"github.com/ulikunitz/xz"
)

func (a *Archive) Unxz() error {
	// https://pkg.go.dev/github.com/ulikunitz/xz#section-readme
	xzr, err := xz.NewReader(a.FileHandle)
	if err != nil {
		return err
	}
	a.StreamHandle = ioutil.NopCloser(xzr)
	return err
}
