package extract

import (
	"io"
	"io/fs"
	"os"

	log "github.com/sirupsen/logrus"
)

type Archive struct {
	Path         string
	PathNoExt    string
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
		PathNoExt:    "",
		FileHandle:   f,
		FileStats:    stats,
		StreamHandle: nil,
	}, nil
}

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
