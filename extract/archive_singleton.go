package extract

import (
	"fmt"
	"io/fs"

	"github.com/backplane/ghlatest/util"
	log "github.com/sirupsen/logrus"
)

// WriteSingleton extracts the Archive's solitary file into the given output
// directory using io.Copy. This is typically applied when a single file is
// compressed with a utility like gzip. If the file to be written conflicts
// with an existing file in the outputDir then extraction will stop unless the
// overwrite argument is set to true.
func (a *Archive) WriteSingleton(outputPath string, mode fs.FileMode, overwrite bool) error {
	if a.StreamHandle == nil {
		// StreamHandle must be available because we (previously) decompressed
		return fmt.Errorf("nil StreamHandle; didn't find any decompressed data")
	}
	log.Debug("using StreamHandle to write output file")

	_, err := util.NewFileFromSource(outputPath, mode, overwrite, a.StreamHandle)

	return err
}
