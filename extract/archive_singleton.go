package extract

import (
	"fmt"
	"io"
	"io/fs"

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
		return fmt.Errorf("StreamHandle is nil; didn't find any decompressed data")
	}
	log.Debug("Using StreamHandle to write output file")

	outputFile, err := NewFile(outputPath, mode, overwrite)
	if err != nil {
		return fmt.Errorf("Failed to open output file \"%s\"; error: %s", outputPath, err)
	}

	_, err = io.Copy(outputFile, a.StreamHandle)
	if err != nil {
		return fmt.Errorf("Failed to write data to output file; error: %s", err)
	}
	outputFile.Close()
	log.Infof("Created %s mode: %#o", outputPath, mode)

	return nil
}
