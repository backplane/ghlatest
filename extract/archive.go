package extract

import (
	"io"
	"io/fs"
	"os"

	log "github.com/sirupsen/logrus"
)

// Archive is a handle for optionally-compressed file archives which contain
// one or more files.
type Archive struct {
	Path         string        // full path to the archive file
	PathNoExt    string        // full path to the archive file with any recognized filename extensions removed
	FileHandle   *os.File      // file handle for the archive file
	FileStats    fs.FileInfo   // file statistics for the archive file
	StreamHandle io.ReadCloser // handle for optional decompression reader
}

// OpenArchive opens the archive file at the given path and returns a handle
// suitable for file extraction operations on that archive
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

// Close handles closing the resources contained in an Archive handle
func (a *Archive) Close() {
	if a.StreamHandle != nil {
		if err := a.StreamHandle.Close(); err != nil {
			log.Fatalln(err)
		}
	}
	if a.FileHandle != nil {
		if err := a.FileHandle.Close(); err != nil {
			log.Fatalln(err)
		}
	}
}
