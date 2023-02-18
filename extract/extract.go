package extract

import (
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type Archive struct {
	Path         string
	FileHandle   *os.File
	FileStats    fs.FileInfo
	StreamHandle io.ReadCloser
}

func OpenArchive(filePath string) (*Archive, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	stats, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &Archive{
		Path:         filePath,
		FileHandle:   f,
		FileStats:    stats,
		StreamHandle: nil,
	}, nil
}

func (a *Archive) Close() {
	if a.StreamHandle != nil {
		if err := a.StreamHandle.Close(); err != nil {
			logrus.Fatalln(err)
		}
	}
	if a.FileHandle != nil {
		if err := a.FileHandle.Close(); err != nil {
			logrus.Fatalln(err)
		}
	}
}

func ExtractFile(filePath string, rawFilters []string, overwrite bool) bool {
	fileName := path.Base(filePath)

	filters := make([]*regexp.Regexp, 0)
	for _, filterStr := range rawFilters {
		filter, err := regexp.Compile(filterStr)
		if err != nil {
			logrus.Fatalf("failed to compile --keep filter: \"%s\" - error: %s", filterStr, err)
		}
		filters = append(filters, filter)
	}

	sevenZipFileName := regexp.MustCompile(`(?i)\.7z$`)
	tarFileName := regexp.MustCompile(`(?i)\.tar$`)
	tbz2FileName := regexp.MustCompile(`(?i)\.(tbz2|tar\.bz2)$`)
	tgzFileName := regexp.MustCompile(`(?i)\.(tgz|tar\.gz)$`)
	txzFileName := regexp.MustCompile(`(?i)\.(txz|tar\.xz)$`)
	zipFileName := regexp.MustCompile(`(?i)\.zip$`)

	a, err := OpenArchive(filePath)
	if err != nil {
		logrus.Fatalln(err)
	}
	defer a.Close()

	switch {
	case sevenZipFileName.MatchString(fileName):
		logrus.Fatalln("7z extraction unimplemented")
	case tarFileName.MatchString(fileName):
		logrus.Infof("un-tarring %s", filePath)
		a.Untar(".", filters, overwrite)
	case tbz2FileName.MatchString(fileName):
		logrus.Infof("uncompressing (bzip2) %s", fileName)
		err = a.Bunzip2()
		if err != nil {
			logrus.Fatalf("uncompressing (bzip2) %s failed", fileName)
		}
		logrus.Infof("un-tarring %s", fileName)
		a.Untar(".", filters, overwrite)
	case tgzFileName.MatchString(fileName):
		logrus.Infof("uncompressing (gzip) %s", fileName)
		err = a.Gunzip()
		if err != nil {
			logrus.Fatalf("uncompressing (gzip) %s failed: %s", fileName, err)
		}
		logrus.Infof("un-tarring %s", fileName)
		a.Untar(".", filters, overwrite)
	case txzFileName.MatchString(fileName):
		logrus.Infof("uncompressing (xz) %s", fileName)
		err = a.Unxz()
		if err != nil {
			logrus.Fatalf("uncompressing (xz) %s failed", fileName)
		}
		logrus.Infof("un-tarring %s", fileName)
		a.Untar(".", filters, overwrite)
	case zipFileName.MatchString(fileName):
		logrus.Infof("unzipping %s\n", filePath)
		a.Unzip(".", filters, overwrite)
	default:
		logrus.Fatalf("%s extraction unimplemented", path.Ext(fileName))
	}
	logrus.Info("extraction complete")

	return false
}

func NormalizeFilePath(path string) string {
	driveLabels := regexp.MustCompile(`^[A-Z]:\\*`)
	onlyDotsAndSpaces := regexp.MustCompile(`^[.\s]*$`)
	validatedPathParts := make([]string, 0)

	normalized := strings.TrimSpace(path)
	normalized = driveLabels.ReplaceAllLiteralString(normalized, ``)
	normalized = strings.ReplaceAll(normalized, `\`, `/`)
	normalized = filepath.Clean(normalized)
	for _, part := range strings.Split(normalized, `/`) {
		trimmed := strings.TrimSpace(part)
		if onlyDotsAndSpaces.MatchString(trimmed) {
			continue
		}
		validatedPathParts = append(validatedPathParts, trimmed)
	}
	normalized = strings.Join(validatedPathParts, "/")
	normalized = filepath.Clean(normalized)

	return normalized
}
