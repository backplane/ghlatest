package extract

import (
	"archive/tar"
	"io"
	"os"

	"github.com/backplane/ghlatest/util"
	log "github.com/sirupsen/logrus"
)

// Untar extracts the Archive's contents into the given output directory using
// a tar file reader. If there are any filters in the given FilterSet then files
// are only extracted if they match one of the given filters. If the files to
// be created conflict with existing files in the outputDir then extraction
// will stop unless the overwrite argument is set to true.
func (a *Archive) Untar(outputDir string, filters util.FilterSet, overwrite bool) []string {
	// see: https://pkg.go.dev/archive/tar#pkg-overview
	// Open and iterate through the files in the archive.

	// fixme: outputDir is not currently implemented!

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
	var filtering bool = len(filters) > 0

	for {
		f, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}

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
		switch f.Typeflag {
		case tar.TypeReg:
			_, err = util.NewFileFromSource(filePath, permissions, overwrite, tr)
			if err != nil {
				log.Errorf("%s: extracting file failed; error: %s; skipping any remaining files in archive", filePath, err)
				goto CONTINUE_OUTER
			}
		case tar.TypeLink:
			err = os.Link(f.Linkname, filePath)
			if err != nil {
				log.Errorf("%s: creating symlink failed; error: %s; skipping any remaining files in archive", filePath, err)
				goto CONTINUE_OUTER
			}
		case tar.TypeSymlink:
			err = os.Symlink(f.Linkname, filePath)
			if err != nil {
				log.Errorf("%s: creating symlink failed; error: %s; skipping any remaining files in archive", filePath, err)
				goto CONTINUE_OUTER
			}
		case tar.TypeDir:
			err := util.NewDirectory(filePath, permissions)
			if err != nil {
				log.Errorf("%s: mkdir failed; error: %s; skipping any remaining files in archive", filePath, err)
				goto CONTINUE_OUTER
			}
		case tar.TypeFifo:
			log.Errorf("%s: mkfifo skipped; extracting FIFOs is not currently supported", filePath)
			continue
		case tar.TypeXGlobalHeader:
			log.Debugf("%s: skipping pax global header file", filePath)
			continue
		case tar.TypeGNUSparse:
			log.Debugf("%s: GNU Sparse files are not supported", filePath)
			goto CONTINUE_OUTER
		default:
			log.Errorf("%s: unknown type %d; skipping any remaining files in archive", filePath, f.Typeflag)
		}
		extractedFiles = append(extractedFiles, filePath)
	}
CONTINUE_OUTER:
	return extractedFiles
}
