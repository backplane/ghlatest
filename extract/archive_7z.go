package extract

import (
	"github.com/backplane/ghlatest/util"
	"github.com/bodgit/sevenzip"
	log "github.com/sirupsen/logrus"
)

// Un7z extracts the Archive's contents into the given output directory using
// the a sevenzip file reader. If there are any filters in the given FilterSet then
// files are only extracted if they match one of the given filters. If the
// files to be created conflict with existing files in the outputDir then
// extraction will stop unless the overwrite argument is set to true.
func (a *Archive) Un7z(outputDir string, filters util.FilterSet, overwrite bool) []string {
	// https://pkg.go.dev/github.com/bodgit/sevenzip@v1.4.0

	// fixme: outputDir is not currently implemented!

	r, err := sevenzip.NewReader(a.FileHandle, a.FileStats.Size())
	if err != nil {
		log.Fatal(err)
	}

	extractedFiles := make([]string, 0)
	var filtering bool = len(filters) > 0

	for _, f := range r.File {
		filePath := util.NormalizeFilePath(f.Name)
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
			err := util.NewDirectory(filePath, mode)
			if err != nil {
				log.Errorf("skipping any remaining files in archive")
				break
			}
			extractedFiles = append(extractedFiles, filePath)
			continue
		}

		srcContents, err := f.Open()
		if err != nil {
			log.Fatalf("Opening source contents of %s failed; error: %s", filePath, err)
		}
		defer srcContents.Close()

		_, err = util.NewFileFromSource(filePath, mode, overwrite, srcContents)
		if err != nil {
			log.Errorf("skipping any remaining files in archive")
			break
		}

		extractedFiles = append(extractedFiles, filePath)
	}
	return extractedFiles
}
