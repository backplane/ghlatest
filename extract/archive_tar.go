package extract

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"path"
	"regexp"

	"github.com/sirupsen/logrus"
)

func (a *Archive) Untar(outputDir string, filters []*regexp.Regexp, overwrite bool) []string {
	// see: https://pkg.go.dev/archive/tar#pkg-overview
	// Open and iterate through the files in the archive.
	var tr *tar.Reader
	if a.StreamHandle != nil {
		// StreamHandle would be available if we're decompressing as well
		logrus.Info("Using StreamHandle to untar")
		tr = tar.NewReader(a.StreamHandle)
	} else {
		logrus.Info("Using FileHandle to untar")
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
			logrus.Fatal(err)
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
			logrus.Infof("creating directory %s mode: %#o", filePath, permissions)
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

		fileMode := os.O_WRONLY | os.O_CREATE
		if overwrite {
			fileMode |= os.O_EXCL
		}

		outputFile, err := os.OpenFile(filePath, fileMode, permissions)
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(outputFile, tr)
		if err != nil {
			log.Fatal(err)
		}
		outputFile.Close()
		logrus.Infof("created %s mode: %#o", filePath, permissions)
		extractedFiles = append(extractedFiles, filePath)

	}
	return extractedFiles
}
