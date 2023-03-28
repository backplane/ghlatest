package util

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// byteCountIEC returns a string describing the givne number of bytes in
// "human-readable" form. It uses IEC multiples (with e.g. 1024 bits to a byte)
func byteCountIEC(b int64) string {
	// courtesy of: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
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

// DownloadFile places the contents of the given url into a local file at the
// given path with the given mode. If the overwrite flag is set then any
// existing files with conflicting names will be overwritten
func DownloadFile(url string, filePath string, mode os.FileMode, overwrite bool) error {
	// generally applicable utility for downloading the contents of a url to
	// a given file path.
	// copied (with minor mod.) from: https://stackoverflow.com/a/33853856

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK HTTP response status: %s", resp.Status)
	}
	log.Debugf("disposition: %s\n", resp.Header["Content-Disposition"])

	_, err = NewFileFromSource(filePath, mode, overwrite, resp.Body)
	return err
}

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

// NewFileFrom wraps os.OpenFile and io.Copy to standardize the way the package
// creates output files. The path argument accepts the file path to the given
// output file, the mode argument accepts an [fs.FileMode] which the file will
// be created with, and overwrite determines whether existing files should be
// overwritten or and error produced, the source provides the data that
// will be written to the new file. The bytes written are returned.
func NewFileFromSource(path string, mode fs.FileMode, overwrite bool, source io.Reader) (int64, error) {
	openFlags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		openFlags &^= os.O_EXCL // remove the "don't overwrite" flag
	}

	outputFile, err := os.OpenFile(path, openFlags, mode)
	if err != nil {
		log.Errorf("Opening output file \"%s\" with flags %s failed; error: %s", path, flagsString(openFlags), err)
		return 0, err
	}

	bytes, err := io.Copy(outputFile, source)
	if err != nil {
		log.Errorf("Writing to output file \"%s\" failed; error: %s", path, err)
		return bytes, err
	}

	log.Infof("created file:\"%s\"; mode:%#o; bytes:%d (%s)", path, mode, bytes, byteCountIEC(int64(bytes)))

	return bytes, nil
}

// NewDirectory wraps os.Mkdir to provide standardized directory creation and
// logging
func NewDirectory(path string, mode fs.FileMode) (err error) {
	err = os.Mkdir(path, mode)
	if err != nil {
		log.Infof("Creating directory \"%s\" failed; error: %s", path, err)
		return
	}

	log.Infof("created directory:\"%s\"; mode:%#o", path, mode)
	return
}
