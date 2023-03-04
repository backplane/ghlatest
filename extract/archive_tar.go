package extract

import (
	"archive/tar"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

// Untar extracts the Archive's contents into the given output directory using
// a tar file reader. If there are any filters in the given FilterSet then files
// are only extracted if they match one of the given filters. If the files to
// be created conflict with existing files in the outputDir then extraction
// will stop unless the overwrite argument is set to true.
func (a *Archive) Untar(outputDir string, filters FilterSet, overwrite bool) []string {
	// see: https://pkg.go.dev/archive/tar#pkg-overview
	// Open and iterate through the files in the archive.
	var tr *tar.Reader
	if a.StreamHandle != nil {
		// StreamHandle would be available if we're decompressing as well
		log.Debug("untar selected StreamHandle")
		tr = tar.NewReader(a.StreamHandle)
	} else {
		log.Debug("untar selected FileHandle")
		tr = tar.NewReader(a.FileHandle)
	}

	extractedFiles := make([]string, 0)
	var filtering bool = false
	if len(filters) > 0 {
		filtering = true
	}

	for {
		f, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}

		filePath := NormalizeFilePath(f.Name)
		if filtering {
			var include_file bool = false
			for _, filter := range filters {
				if filter.MatchString(filePath) {
					include_file = true
					break
				}
			}
			if !include_file {
				continue
			}
		}

		permissions := f.FileInfo().Mode().Perm()
		if f.FileInfo().IsDir() {
			log.Infof("creating directory %s mode: %#o", filePath, permissions)
			err := os.MkdirAll(filePath, permissions)
			if err != nil {
				log.Fatal(err)
			}
			extractedFiles = append(extractedFiles, filePath)
			continue
		}

		outputFile, err := NewFile(filePath, permissions, overwrite)
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(outputFile, tr)
		if err != nil {
			log.Fatal(err)
		}
		outputFile.Close()
		log.Infof("created %s mode: %#o", filePath, permissions)
		extractedFiles = append(extractedFiles, filePath)

	}
	return extractedFiles
}
