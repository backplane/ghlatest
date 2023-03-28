package extract

import (
	"archive/zip"
	"os"

	"github.com/backplane/ghlatest/util"
	log "github.com/sirupsen/logrus"
)

// Unzip extracts the Archive's contents into the given output directory using
// the a zip file reader. If there are any filters in the given FilterSet then
// files are only extracted if they match one of the given filters. If the
// files to be created conflict with existing files in the outputDir then
// extraction will stop unless the overwrite argument is set to true.
func (a *Archive) Unzip(outputDir string, filters util.FilterSet, overwrite bool) []string {
	// https://pkg.go.dev/archive/zip@go1.20.1#example-Reader
	// Open a zip archive for reading.

	// fixme: outputDir is not currently implemented!

	r, err := zip.NewReader(a.FileHandle, a.FileStats.Size())
	if err != nil {
		log.Fatal(err)
	}

	extractedFiles := make([]string, 0)
	var filtering bool = len(filters) > 0

	for _, f := range r.File {
		filePath := util.NormalizeFilePath(f.Name)
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

		permissions := f.FileInfo().Mode().Perm()
		fType := f.FileInfo().Mode().Type()
		switch {
		case fType.IsRegular():
			contents, err := f.Open()
			if err != nil {
				log.Fatalf("%s: opening source contents of failed; error: %s", filePath, err)
			}
			defer contents.Close()

			_, err = util.NewFileFromSource(filePath, permissions, overwrite, contents)
			if err != nil {
				log.Errorf("%s: extracting file failed; error: %s; skipping any remaining files in archive", filePath, err)
				goto CONTINUE_OUTER
			}
		case fType.IsDir():
			err := util.NewDirectory(filePath, permissions)
			if err != nil {
				log.Errorf("%s: mkdir failed; error: %s; skipping any remaining files in archive", filePath, err)
				goto CONTINUE_OUTER
			}
		case fType&os.ModeSymlink != 0:
			log.Errorf("%s: symlink skipped; extracting symlinks is not currently supported", filePath)
			continue
		case fType&os.ModeNamedPipe != 0:
			log.Errorf("%s: mkfifo skipped; extracting FIFOs is not currently supported", filePath)
			continue
		default:
			log.Errorf("%s: unknown type; skipping any remaining files in archive", filePath)
			goto CONTINUE_OUTER
		}
		extractedFiles = append(extractedFiles, filePath)
	}
CONTINUE_OUTER:
	return extractedFiles
}
