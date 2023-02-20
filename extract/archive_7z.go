package extract

import (
	"io"
	"os"

	"github.com/bodgit/sevenzip"
	log "github.com/sirupsen/logrus"
)

// Un7z extracts the Archive's contents into the given output directory using
// the a sevenzip file reader. If there are any filters in the given FilterSet then
// files are only extracted if they match one of the given filters. If the
// files to be created conflict with existing files in the outputDir then
// extraction will stop unless the overwrite argument is set to true.
func (a *Archive) Un7z(outputDir string, filters FilterSet, overwrite bool) []string {
	// https://pkg.go.dev/github.com/bodgit/sevenzip@v1.4.0

	// fixme: outputDir is not currently implemented!

	r, err := sevenzip.NewReader(a.FileHandle, a.FileStats.Size())
	if err != nil {
		log.Fatal(err)
	}

	extractedFiles := make([]string, 0)
	var filtering bool = len(filters) > 0

	for _, f := range r.File {
		filePath := NormalizeFilePath(f.Name)
		mode := f.Mode().Perm()
		if filtering {
			var include_file bool = false
			for _, filter := range filters {
				if filter.MatchString(filePath) {
					include_file = true
					break
				}
			}
			if !include_file {
				log.Debugf("Skipping %s", filePath)
				continue
			}
		}

		if f.FileInfo().IsDir() {
			log.Infof("creating directory %s mode: %#o", filePath, mode)
			if err := os.Mkdir(filePath, mode); err != nil {
				log.Fatal(err)
			}
			extractedFiles = append(extractedFiles, filePath)
			continue
		}

		srcContents, err := f.Open()
		if err != nil {
			log.Fatalf("Opening source contents of %s failed; error: %s", filePath, err)
		}

		outputFile, err := NewFile(filePath, mode, overwrite)
		if err != nil {
			log.Fatalf("Opening output file of %s failed; error: %s", filePath, err)
		}

		bytes, err := io.Copy(outputFile, srcContents)
		if err != nil {
			log.Fatal(err)
		}
		outputFile.Close()
		srcContents.Close()
		log.Infof("created %d-byte file: %s mode: %#o", bytes, filePath, mode)
		extractedFiles = append(extractedFiles, filePath)
	}
	return extractedFiles
}
