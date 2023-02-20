// Extract implements the extraction of files from various file archive or compression formats.
//
// The entrypoint for the package is ExtractFile which is meant to call the
// appropriate handlers for the file at the given path based on the filename
// extensions. OpenArchive may also be used. The resulting Archive struct has
// methods which handle common archive types.
package extract

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// FilterSet contains compiled regular expressions which identify files to be
// included during an extraction process. If the FilterSet is empty all files
// will be extracted.
type FilterSet []*regexp.Regexp

// HandlerFunc is the signature of functions which are called by ExtractFile to
// process an Archive. outputPath is the path to the directory where files
// should be extracted. filters (if non-empty) contain regular expressions
// which select files to be included in the output. When overwrite is true
// existing files may be overwritten by the extraction process (instead of
// halting the program)
type handlerFunc func(a *Archive, outputPath string, filters FilterSet, overwrite bool) error

// ExtHandler is used by ExtractFile to define a regexp matcher for filenames
// and which function can handle extraction for that type of file.
type extHandler struct {
	Matcher *regexp.Regexp
	Handler handlerFunc
}

var extHandlers = []extHandler{
	{regexp.MustCompile(`(?i)\.7z$`), handle7z},
	{regexp.MustCompile(`(?i)\.tar$`), handleTar},
	{regexp.MustCompile(`(?i)\.zip$`), handleZip},
	{regexp.MustCompile(`(?i)\.(tbz2|tar\.bz2)$`), handleTbz2},
	{regexp.MustCompile(`(?i)\.(tgz|tar\.gz)$`), handleTgz},
	{regexp.MustCompile(`(?i)\.(txz|tar\.xz)$`), handleTxz},
	{regexp.MustCompile(`(?i)\.bz2$`), handleBz2},
	{regexp.MustCompile(`(?i)\.gz$`), handleGz},
	{regexp.MustCompile(`(?i)\.xz$`), handleXz},
}

// ExtractFile extracts the contents of the file archive at the given filePath.
// The filterStrings argument accepts a slice of strings (which will be compiled
// into [regexp.Regexp] objects) to filter what will be extracted from the file
// archive.
func ExtractFile(filePath string, filterStrings []string, overwrite bool) error {
	filters, err := compileFilters(filterStrings)
	if err != nil {
		log.Fatalf("failed to compile --keep filters, error: %s", err)
	}

	a, err := OpenArchive(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer a.Close()

	for _, h := range extHandlers {
		if !h.Matcher.MatchString(filePath) {
			continue
		}
		// it's important to populate this with the first matcher because we
		// need to support different handlers for example.tar.gz and example.gz
		a.PathNoExt = h.Matcher.ReplaceAllString(filePath, "")
		if err := h.Handler(a, ".", filters, overwrite); err != nil {
			log.Fatalf("failed to handle extraction for %s; error: %s", filePath, err)
		}
		log.Info("extraction complete")
		return nil
	}

	return fmt.Errorf("Don't know how to extract %s (no handler)", filePath)
}

func handle7z(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	return fmt.Errorf("Cannot extract %s -- 7z extraction not yet implemented", a.Path)
}

func handleTar(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	log.Infof("extracting (tar) %s", a.Path)
	extractedFiles := a.Untar(outputPath, filters, overwrite)
	if len(extractedFiles) > 0 {
		return nil
	}
	return fmt.Errorf("problem extracting %s: no files were produced", a.Path)
}

func handleZip(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	log.Infof("extracting (zip) %s", a.Path)
	extractedFiles := a.Unzip(outputPath, filters, overwrite)
	if len(extractedFiles) > 0 {
		return nil
	}
	return fmt.Errorf("problem extracting %s: no files were produced", a.Path)
}

func handleTbz2(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	log.Infof("extracting (tbz2) %s", a.Path)
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
	log.Infof("extracting (tgz) %s", a.Path)
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
	log.Infof("extracting (txz) %s", a.Path)
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
	log.Infof("extracting (bz2) %s", a.Path)
	if err := a.Bunzip2(); err != nil {
		return fmt.Errorf("uncompressing (bz2) %s failed; error: %s", a.Path, err)
	}
	if err := a.WriteSingleton(a.PathNoExt, a.FileStats.Mode().Perm(), overwrite); err != nil {
		return fmt.Errorf("Unable to write the downloaded file %s; error: %s", a.PathNoExt, err)
	}
	return nil
}

func handleGz(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	log.Infof("extracting (gz) %s", a.Path)
	if err := a.Gunzip(); err != nil {
		return fmt.Errorf("uncompressing (gz) %s failed; error: %s", a.Path, err)
	}
	if err := a.WriteSingleton(a.PathNoExt, a.FileStats.Mode().Perm(), overwrite); err != nil {
		return fmt.Errorf("Unable to write the downloaded file %s; error: %s", a.PathNoExt, err)
	}
	return nil
}

func handleXz(a *Archive, outputPath string, filters FilterSet, overwrite bool) error {
	log.Infof("extracting (xz) %s", a.Path)
	if err := a.Unxz(); err != nil {
		return fmt.Errorf("uncompressing (xz) %s failed; error: %s", a.Path, err)
	}
	if err := a.WriteSingleton(a.PathNoExt, a.FileStats.Mode().Perm(), overwrite); err != nil {
		return fmt.Errorf("Unable to write the downloaded file %s; error: %s", a.PathNoExt, err)
	}
	return nil
}
