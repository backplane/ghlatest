package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/backplane/ghlatest/extract"
	log "github.com/sirupsen/logrus"
)

func downloadFile(url string, filePath string, mode os.FileMode, overwrite bool) error {
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

	// open the output file
	outputFH, err := extract.NewFile(filePath, mode, overwrite)
	if err != nil {
		return fmt.Errorf("couldn't open '%s' for writing. Error: %v", filePath, err)
	}
	defer outputFH.Close()

	// write the body to file
	bytes, err := io.Copy(outputFH, resp.Body)
	if err != nil {
		return fmt.Errorf("couldn't copy download data into output file '%v'. Error: %v", outputFH, err)
	}
	log.Infof("wrote %d bytes to %s", bytes, filePath)

	return nil
}
