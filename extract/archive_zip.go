package extract

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
)

func (a *Archive) Unzip(outputDir string, filters []*regexp.Regexp, overwrite bool) []string {
	// https://pkg.go.dev/archive/zip@go1.20.1#example-Reader
	// Open a zip archive for reading.

	r, err := zip.NewReader(a.FileHandle, a.FileStats.Size())
	if err != nil {
		log.Fatal(err)
	}

	extractedFiles := make([]string, 0)
	var filtering bool = false
	if len(filters) > 0 {
		filtering = true
	}

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
				logrus.Debugf("Skipping %s", filePath)
				continue
			}
		}

		if f.FileInfo().IsDir() {
			logrus.Infof("creating directory %s mode: %#o", filePath, mode)
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

		openFlags := os.O_WRONLY | os.O_CREATE
		if overwrite {
			openFlags |= os.O_EXCL
		}
		outputFile, err := os.OpenFile(filePath, openFlags, mode)
		if err != nil {
			log.Fatalf("Opening output file of %s failed; error: %s", filePath, err)
		}

		bytes, err := io.Copy(outputFile, srcContents)
		if err != nil {
			log.Fatal(err)
		}
		outputFile.Close()
		srcContents.Close()
		logrus.Infof("created %d-byte file: %s mode: %#o", bytes, filePath, mode)
		extractedFiles = append(extractedFiles, filePath)
	}
	return extractedFiles
}
