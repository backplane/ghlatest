package extract

import (
	"fmt"
	"io"
	"io/fs"

	log "github.com/sirupsen/logrus"
)

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
