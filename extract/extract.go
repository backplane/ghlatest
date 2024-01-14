// Extract implements the extraction of files from various file archive or compression formats.
//
// The entrypoint for the package is ExtractFile which is meant to call the
// appropriate handlers for the file at the given path based on the filename
// extensions. OpenArchive may also be used. The resulting Archive struct has
// methods which handle common archive types.
package extract

// this file is an alternative implementation of the code in extract

import (
	"fmt"
	"regexp"

	"github.com/backplane/ghlatest/util"
	log "github.com/sirupsen/logrus"
)

// aop is an Archive operation
type aop int

// <100 = decompressors that don't write files
const (
	opUnbzip2 aop = iota
	opUngzip
	opUnxz
)

// >= 100 = single-file writers
const (
	opWriteSingleton aop = iota + 100
)

// >= 200 = multi-file writers
const (
	opUn7z aop = iota + 200
	opUntar
	opUnzip
)

var archiveOpNames = map[aop]string{
	opUn7z:           "7z",
	opUnbzip2:        "bzip2",
	opUngzip:         "gzip",
	opUntar:          "tar",
	opUnxz:           "xz",
	opUnzip:          "zip",
	opWriteSingleton: "file",
}

type filenameStrategy struct {
	FilenameRegexp *regexp.Regexp
	Operations     []aop
}

var strategies = []filenameStrategy{
	{regexp.MustCompile(`(?i)\.7z$`), []aop{opUn7z}},
	{regexp.MustCompile(`(?i)\.tar$`), []aop{opUntar}},
	{regexp.MustCompile(`(?i)\.zip$`), []aop{opUnzip}},
	{regexp.MustCompile(`(?i)\.(tbz2|tar\.bz2)$`), []aop{opUnbzip2, opUntar}},
	{regexp.MustCompile(`(?i)\.(tgz|tar\.gz)$`), []aop{opUngzip, opUntar}},
	{regexp.MustCompile(`(?i)\.(txz|tar\.xz)$`), []aop{opUnxz, opUntar}},
	{regexp.MustCompile(`(?i)\.bz2$`), []aop{opUnbzip2, opWriteSingleton}},
	{regexp.MustCompile(`(?i)\.gz$`), []aop{opUngzip, opWriteSingleton}},
	{regexp.MustCompile(`(?i)\.xz$`), []aop{opUnxz, opWriteSingleton}},
}

// ExtractFile extracts the contents of the file archive at the given filePath.
// The filterStrings argument accepts a slice of strings (which will be compiled
// into [regexp.Regexp] objects) to filter what will be extracted from the file
// archive.
func ExtractFile(filePath string, filterStrings []string, overwrite bool) error {
	filters, err := util.CompileFilters(filterStrings)
	if err != nil {
		log.Fatalf("failed to compile --keep filters, error: %s", err)
	}

	a, err := OpenArchive(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer a.Close()

	var outputDir string = "."

	for _, strategy := range strategies {
		if !strategy.FilenameRegexp.MatchString(filePath) {
			continue
		}
		// it's important to populate this with the first matcher found because we
		// need to support different handlers for example.tar.gz and example.gz
		a.PathNoExt = strategy.FilenameRegexp.ReplaceAllString(filePath, "")
		for _, op := range strategy.Operations {
			var err error
			var extractedFiles []string

			// op >= 200: multi-file writers, which always terminate the operation list
			if op >= 200 {
				log.Infof("extracting (%s) %s", archiveOpNames[op], a.Path)
				switch op {
				case opUn7z:
					extractedFiles = a.Un7z(outputDir, filters, overwrite)
				case opUntar:
					extractedFiles = a.Untar(outputDir, filters, overwrite)
				case opUnzip:
					extractedFiles = a.Unzip(outputDir, filters, overwrite)
				default:
					panic(fmt.Sprintf("Encountered unhandled archive operation %d (%s)", op, archiveOpNames[op]))
				}
				if len(extractedFiles) < 1 {
					return fmt.Errorf("no files were extracted from archive; stopping extraction")
				}
				return nil
			}

			// op >= 100: single-file writers which always terminate the operation list
			if op >= 100 {
				log.Debugf("writing decompressed contents of %s", a.Path)
				switch op {
				case opWriteSingleton:
					err = a.WriteSingleton(a.PathNoExt, a.FileStats.Mode().Perm(), overwrite)
				default:
					panic(fmt.Sprintf("Encountered unhandled archive operation %d (%s)", op, archiveOpNames[op]))
				}
				return err
			}

			// op < 100: decompressors that don't write files and never terminate the operation list
			log.Infof("decompressing (%s) %s", archiveOpNames[op], a.Path)
			switch op {
			case opUnbzip2:
				err = a.Unbzip2()
			case opUngzip:
				err = a.Ungzip()
			case opUnxz:
				err = a.Unxz()
			default:
				panic(fmt.Sprintf("Encountered unhandled archive operation %d (%s)", op, archiveOpNames[op]))
			}
			if err != nil {
				return fmt.Errorf(`decompressing (%s) "%s" failed; error:%s`, archiveOpNames[op], a.Path, err)
			}
			continue
		}

		// none of the operations returned from this function
		panic("reached code that should be unreachable oplists should terminate with ops >=100 but were're still here")
	}

	// none of the matchers matched
	return fmt.Errorf(`don't know how to extract file:"%s"`, filePath)
}
