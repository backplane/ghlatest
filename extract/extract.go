package extract

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type Archive struct {
	Path         string
	PathNoExt    string
	FileHandle   *os.File
	FileStats    fs.FileInfo
	StreamHandle io.ReadCloser
}

type FilterSet []*regexp.Regexp

type HandlerSpec struct {
	Matcher *regexp.Regexp
	Handler func(a *Archive, outputPath string, filters FilterSet, overwrite bool) error
}

type HandlersList []HandlerSpec

var Handlers HandlersList = HandlersList{
	HandlerSpec{regexp.MustCompile(`(?i)\.7z$`), handle7z},
	HandlerSpec{regexp.MustCompile(`(?i)\.tar$`), handleTar},
	HandlerSpec{regexp.MustCompile(`(?i)\.zip$`), handleZip},
	HandlerSpec{regexp.MustCompile(`(?i)\.(tbz2|tar\.bz2)$`), handleTbz2},
	HandlerSpec{regexp.MustCompile(`(?i)\.(tgz|tar\.gz)$`), handleTgz},
	HandlerSpec{regexp.MustCompile(`(?i)\.(txz|tar\.xz)$`), handleTxz},
	HandlerSpec{regexp.MustCompile(`(?i)\.bz2$`), handleBz2},
	HandlerSpec{regexp.MustCompile(`(?i)\.gz$`), handleGz},
	HandlerSpec{regexp.MustCompile(`(?i)\.xz$`), handleXz},
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
		PathNoExt:    "",
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

func ExtractFile(filePath string, rawFilters []string, overwrite bool) error {
	filters := make(FilterSet, 0)
	for _, filterStr := range rawFilters {
		filter, err := regexp.Compile(filterStr)
		if err != nil {
			logrus.Fatalf("failed to compile --keep filter: \"%s\" - error: %s", filterStr, err)
		}
		filters = append(filters, filter)
	}

	a, err := OpenArchive(filePath)
	if err != nil {
		logrus.Fatalln(err)
	}
	defer a.Close()

	for _, h := range Handlers {
		if !h.Matcher.MatchString(filePath) {
			continue
		}
		// it's important to populate this with the first matcher because we
		// need to support different handlers for example.tar.gz and example.gz
		a.PathNoExt = h.Matcher.ReplaceAllString(filePath, "")
		if err := h.Handler(a, ".", filters, overwrite); err != nil {
			logrus.Fatalf("failed to handle extraction for %s; error: %s", filePath, err)
		}
		logrus.Info("extraction complete")
		return nil
	}

	return fmt.Errorf("Don't know how to extract %s (no handler)", filePath)
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

func handle7z(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	return fmt.Errorf("Cannot extract %s -- 7z extraction not yet unimplemented", a.Path)
}

func handleTar(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	logrus.Infof("extracting (tar) %s", a.Path)
	extractedFiles := a.Untar(outputPath, filters, overwrite)
	if len(extractedFiles) > 0 {
		return nil
	}
	return fmt.Errorf("problem extracting %s: no files were produced", a.Path)
}

func handleZip(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	logrus.Infof("extracting (zip) %s", a.Path)
	extractedFiles := a.Unzip(outputPath, filters, overwrite)
	if len(extractedFiles) > 0 {
		return nil
	}
	return fmt.Errorf("problem extracting %s: no files were produced", a.Path)
}

func handleTbz2(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	logrus.Infof("extracting (tbz2) %s", a.Path)
	if err := a.Bunzip2(); err != nil {
		return fmt.Errorf("uncompressing (bzip2) %s failed; error: %s", a.Path, err)
	}
	extractedFiles := a.Untar(outputPath, filters, overwrite)
	if len(extractedFiles) > 0 {
		return nil
	}
	return fmt.Errorf("problem extracting %s: no files were produced", a.Path)
}

func handleTgz(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	logrus.Infof("extracting (tgz) %s", a.Path)
	if err := a.Gunzip(); err != nil {
		return fmt.Errorf("uncompressing (gzip) %s failed; error: %s", a.Path, err)
	}
	extractedFiles := a.Untar(outputPath, filters, overwrite)
	if len(extractedFiles) > 0 {
		return nil
	}
	return fmt.Errorf("problem extracting %s: no files were produced", a.Path)
}

func handleTxz(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	logrus.Infof("extracting (txz) %s", a.Path)
	if err := a.Unxz(); err != nil {
		return fmt.Errorf("uncompressing (xz) %s failed; error: %s", a.Path, err)
	}
	extractedFiles := a.Untar(outputPath, filters, overwrite)
	if len(extractedFiles) > 0 {
		return nil
	}
	return fmt.Errorf("problem extracting %s: no files were produced", a.Path)
}

func handleBz2(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	logrus.Infof("extracting (bz2) %s", a.Path)
	if err := a.Bunzip2(); err != nil {
		return fmt.Errorf("uncompressing (bz2) %s failed; error: %s", a.Path, err)
	}
	if err := a.WriteSingleton(a.PathNoExt, a.FileStats.Mode().Perm(), overwrite); err != nil {
		return fmt.Errorf("Unable to write the downloaded file %s; error: %s", a.PathNoExt, err)
	}
	return nil
}

func handleGz(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	logrus.Infof("extracting (gz) %s", a.Path)
	if err := a.Gunzip(); err != nil {
		return fmt.Errorf("uncompressing (gz) %s failed; error: %s", a.Path, err)
	}
	if err := a.WriteSingleton(a.PathNoExt, a.FileStats.Mode().Perm(), overwrite); err != nil {
		return fmt.Errorf("Unable to write the downloaded file %s; error: %s", a.PathNoExt, err)
	}
	return nil
}

func handleXz(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	logrus.Infof("extracting (xz) %s", a.Path)
	if err := a.Unxz(); err != nil {
		return fmt.Errorf("uncompressing (xz) %s failed; error: %s", a.Path, err)
	}
	if err := a.WriteSingleton(a.PathNoExt, a.FileStats.Mode().Perm(), overwrite); err != nil {
		return fmt.Errorf("Unable to write the downloaded file %s; error: %s", a.PathNoExt, err)
	}
	return nil
}
