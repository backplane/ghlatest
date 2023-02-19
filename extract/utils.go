package extract

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

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

func NewFile(path string, mode fs.FileMode, overwrite bool) (*os.File, error) {
	openFlags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		// clear the exclusive flag
		openFlags &^= os.O_EXCL
	}

	log.Debugf("Creating new file; path:%s; flags:%v; mode:%#o", path, FlagsString(openFlags), mode)
	return os.OpenFile(path, openFlags, mode)
}

func FlagsString(flags int) string {
	flagDefs := &[]struct {
		flagValue int
		flagName  string
	}{
		{syscall.O_RDONLY, "O_RDONLY"}, // open the file read-only.
		{syscall.O_WRONLY, "O_WRONLY"}, // open the file write-only.
		{syscall.O_RDWR, "O_RDWR"},     // open the file read-write.
		{syscall.O_APPEND, "O_APPEND"}, // append data to the file when writing.
		{syscall.O_CREAT, "O_CREATE"},  // create a new file if none exists.
		{syscall.O_EXCL, "O_EXCL"},     // used with O_CREATE, file must not exist.
		{syscall.O_SYNC, "O_SYNC"},     // open for synchronous I/O.
		{syscall.O_TRUNC, "O_TRUNC"},   // truncate regular writable file when opened.
	}

	enabledFlags := make([]string, 0, 3)
	for _, m := range *flagDefs {
		if (flags & m.flagValue) != 0 {
			enabledFlags = append(enabledFlags, m.flagName)
		}
	}

	return strings.Join(enabledFlags, "|")
}
