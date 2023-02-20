package extract

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// NormalizeFilePath is a utility function that rewrites file paths specified
// to ensure that they are relative to the current working directory and don't
// have names that are potentially a nuisence for users (such as names composed
// only of whitespace characters, only dots, etc.)
func NormalizeFilePath(path string) string {
	// fixme: this is a mess
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

// NewFile wraps os.OpenFile to standardize the way the package creates output files.
// The path argument accepts the file path to the given output file, the mode
// argument accepts an [fs.FileMode] which the file will be created with, and
// overwrite determines whether existing files should be overwritten or and error
// produced.
func NewFile(path string, mode fs.FileMode, overwrite bool) (*os.File, error) {
	openFlags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		openFlags &^= os.O_EXCL // remove the flag
	}

	log.Debugf("Creating new file; path:%s; flags:%v; mode:%#o", path, flagsString(openFlags), mode)
	return os.OpenFile(path, openFlags, mode)
}

// flagsString returns a string representing the active bits in the given
// [os.Open] flags bitmask
func flagsString(flags int) string {
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

// compileFilters takes a list of regular expression strings and compiles them into a [FilterSet].
func compileFilters(filterStrings []string) (FilterSet, error) {
	filters := make(FilterSet, 0, len(filterStrings))
	for _, filterStr := range filterStrings {
		filter, err := regexp.Compile(filterStr)
		if err != nil {
			return nil, fmt.Errorf("failed to compile filter: \"%s\" - error: %s", filterStr, err)
		}
		filters = append(filters, filter)
	}
	return filters, nil
}
