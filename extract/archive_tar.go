package extract

import (
	"archive/tar"
	"io"
	"os"
	"path"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func (a *Archive) Untar(outputDir string, filters []*regexp.Regexp, overwrite bool) []string {
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
		fileDir, _ := path.Split(filePath)
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
		if fileDir != "" {
			err := os.MkdirAll(filePath, permissions)
			if err != nil {
				log.Fatal(err)
			}
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
