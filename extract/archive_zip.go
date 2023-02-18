package extract

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path"
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

		if f.FileInfo().IsDir() {
			logrus.Infof("creating directory %s mode: %#o", filePath, f.Mode())
			err := os.MkdirAll(filePath, f.Mode())
			if err != nil {
				log.Fatal(err)
			}
			extractedFiles = append(extractedFiles, filePath)
			continue
		}
		if fileDir != "" {
			err := os.MkdirAll(filePath, f.Mode())
			if err != nil {
				log.Fatal(err)
			}
		}

		srcContents, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		fileMode := os.O_WRONLY | os.O_CREATE
		if overwrite {
			fileMode |= os.O_EXCL
		}

		outputFile, err := os.OpenFile(filePath, fileMode, f.Mode())
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(outputFile, srcContents)
		if err != nil {
			log.Fatal(err)
		}
		outputFile.Close()
		srcContents.Close()
		logrus.Infof("created %s mode: %#o", filePath, f.Mode())
		extractedFiles = append(extractedFiles, filePath)
	}
	return extractedFiles
}
