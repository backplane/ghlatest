package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadFile(url string, filepath string, mode os.FileMode, overwrite bool) error {
	// generally applicable utility for downloading the contents of a url to
	// a given file path.
	// copied (with minor mod.) from: https://stackoverflow.com/a/33853856

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("disposition: %s\n", resp.Header["Content-Disposition"])

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// open the output file
	openFlags := os.O_WRONLY | os.O_CREATE
	if overwrite {
		openFlags |= os.O_EXCL
	}

	outputFH, err := os.OpenFile(filepath, openFlags, mode)
	if err != nil {
		return fmt.Errorf("couldn't open '%s' for writing. Error: %v", filepath, err)
	}

	// write the body to file
	_, err = io.Copy(outputFH, resp.Body)
	if err != nil {
		return fmt.Errorf("couldn't copy download data to output file '%v'. Error: %v", outputFH, err)
	}

	// cleanup
	outputFH.Close()

	return nil
}
