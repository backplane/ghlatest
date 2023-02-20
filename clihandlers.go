package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/backplane/ghlatest/extract"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

func getFilterList(c *cli.Context) []*regexp.Regexp {
	filters := make([]*regexp.Regexp, 0)
	var filterString string

	// process the --filter and --ifilter argument
	for _, filterString = range c.StringSlice("filter") {
		filters = append(filters, regexp.MustCompile(filterString))
	}
	for _, filterString = range c.StringSlice("ifilter") {
		filters = append(filters, regexp.MustCompile(`(?i)`+c.String(filterString)))
	}

	// process the --current-arch flag
	if c.Bool("current-arch") {
		filters = append(filters, archRegexp)
	}

	// process the --current-os flag
	if c.Bool("current-os") {
		filters = append(filters, osRegexp)
	}

	return filters
}

func jsonHandler(c *cli.Context) error {
	// extract the owner and repo names from the given URL argument
	if c.NArg() != 1 {
		return fmt.Errorf("you must supply a repo URL argument")
	}
	owner, repo, err := repoURLInfo(c.Args().Get(0))
	if err != nil {
		return err
	}

	// get the json data from the API endpoint
	jsonURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	doc, err := httpContents(jsonURL)
	if err != nil {
		return err
	}

	// indent the json data
	var idoc bytes.Buffer
	if err := json.Indent(&idoc, doc, "", "  "); err != nil {
		return fmt.Errorf("JSON indenting error: %s", err)
	}

	// write the json to stdout
	idoc.WriteTo(os.Stdout)
	fmt.Println()

	return nil
}

func listHandler(c *cli.Context) error {
	// process the repoURL argument
	// make sure the URL looks OK and extract the owner and repo from it
	// fixme: I hate using this as a list... is there a better way?
	if c.NArg() != 1 {
		return fmt.Errorf("you must supply a repo URL argument")
	}
	owner, repo, err := repoURLInfo(c.Args().Get(0))
	if err != nil {
		return err
	}

	for _, assetURL := range latestReleasedAssets(owner, repo, getFilterList(c)) {
		fmt.Println(assetURL)
	}

	return nil
}

func downloadHandler(c *cli.Context) error {
	// process the repoURL argument
	// make sure the URL looks OK and extract the owner and repo from it
	// fixme: I hate using this as a list... is there a better way?
	if c.NArg() != 1 {
		return fmt.Errorf("you must supply a repo URL argument")
	}
	owner, repo, err := repoURLInfo(c.Args().Get(0))
	if err != nil {
		return err
	}

	// determine the assetsURL
	assets := latestReleasedAssets(owner, repo, getFilterList(c))
	if len(assets) != 1 {
		log.Fatalf("found %d matching downloads, use a -f flag to get the match count down to exactly 1\n", len(assets))
	}
	assetURL := assets[0]

	// process the optional output path argument
	var outputpath string
	if c.String("outputpath") != "" {
		outputpath = c.String("outputpath")
		// fixme: should we also validate the given outputpath, like below?
		// we'd have to adjust the regexp to account for file paths
	} else {
		// give it the name from the url -- everything after the last slash
		// kind of like basename
		outputpath = assetURL[strings.LastIndex(assetURL, "/")+1:]
		// quick validation for the above calculated name
		if !filenameRegexp.MatchString(outputpath) {
			return fmt.Errorf("could not correctly calculate an output filename from %s", assetURL)
		}
	}

	// process the mode argument
	// fixme: consider making this a function and adding support for symbolic modes
	mode, err := strconv.ParseUint(c.String("mode"), 8, 32)
	if err != nil {
		return fmt.Errorf("could not process given mode string %s", c.String("mode"))
	}

	// do the download
	err = downloadFile(assets[0], outputpath, os.FileMode(mode), c.Bool("overwrite"))
	if err != nil {
		return err
	}

	// unpack the download
	if c.Bool("extract") {
		extract.ExtractFile(outputpath, c.StringSlice("keep"), c.Bool("overwrite"))
	}

	// cleanup the download
	if c.Bool("remove-archive") {
		if !c.Bool("extract") {
			log.Fatalf("the remove-archive option doesn't make sense unless you also specify extract")
		}

		if err = os.Remove(outputpath); err != nil {
			log.Fatalf("failed to remove the downloaded archive \"%s\", error: %s", outputpath, err)
		}
		log.Infof("removed \"%s\" after extraction", outputpath)
	}

	return nil
}

func extractHandler(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("you must supply a file to extract")
	}
	archivePath := c.Args().Get(0)

	if err := extract.ExtractFile(archivePath, c.StringSlice("keep"), c.Bool("overwrite")); err != nil {
		log.Errorf("failed to extract the archive \"%s\"; error: %s", archivePath, err)
		return err
	}

	// cleanup the download
	if c.Bool("remove-archive") {
		if err := os.Remove(archivePath); err != nil {
			log.Errorf("failed to remove the archive \"%s\", error: %s", archivePath, err)
			return err
		}
		log.Infof("removed \"%s\" after extraction", archivePath)
	}

	return nil
}
