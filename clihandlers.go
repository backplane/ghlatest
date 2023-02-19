package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/backplane/ghlatest/extract"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
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
	err = json.Indent(&idoc, doc, "", "  ")
	if err != nil {
		logrus.Println("JSON indenting error: ", err)
		return err
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
		fmt.Printf("%s\n", assetURL)
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
		logrus.Fatalf("found %d matching downloads, use a -f flag to get the match count down to exactly 1\n", len(assets))
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
	fmt.Printf("wrote to '%s'\n", outputpath)

	if c.Bool("extract") {
		extract.ExtractFile(outputpath, c.StringSlice("keep"), c.Bool("overwrite"))
	}

	if c.Bool("rm") {
		if !c.Bool("extract") {
			log.Fatalf("The --rm option doesn't make sense unless you --extract")
		}

		if err = os.Remove(outputpath); err != nil {
			logrus.Fatalf("failed to --rm the downloaded archive \"%s\", error: %s", outputpath, err)
		}
		logrus.Infof("Removed \"%s\" after extraction", outputpath)
	}

	return nil
}
