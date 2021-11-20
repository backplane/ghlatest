package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	// PROG is the name of this program
	PROG = `ghlatest`

	repoRegexpStr     = `^(?:https://github.com/)?(?P<owner>[^/]+)/(?P<repo>[^/]+)`
	filenameRegexpStr = `^[A-Za-z0-9\_\-\.]{1,256}$`
)

var (
	// version, commit, date, builtBy are provided by goreleaser during build
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"

	repoRegexp     = regexp.MustCompile(repoRegexpStr)
	filenameRegexp = regexp.MustCompile(filenameRegexpStr)
)

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version %s, commit %s, built at %s by %s\n", version, commit, date, builtBy)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = PROG
	app.Version = version
	app.Usage = "Release locator for software on github"
	app.Flags = []cli.Flag{}
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l", "ls"},
			Usage:   "list available releases",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "filter, f",
					Value: "^",
					Usage: "Filter release assets with the given regular expression",
				},
				cli.BoolFlag{
					Name:  "source, s",
					Usage: "List/download source zip files instead of released assets",
				},
			},
			Action: listHandler,
		},
		{
			Name:    "download",
			Aliases: []string{"d", "dl"},
			Usage:   "download the latest available release",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "filter, f",
					Value: "^",
					Usage: "Filter release assets with the given regular expression",
				},
				cli.BoolFlag{
					Name:  "source, s",
					Usage: "List/download source zip files instead of released assets",
				},
				cli.StringFlag{
					Name:  "outputpath, o",
					Usage: "The name of the file to write to",
				},
				cli.StringFlag{
					Name:  "mode, m",
					Value: "0755",
					Usage: "Set the output file's protection mode (ala chmod)",
				},
				cli.BoolFlag{
					Name:  "extract, x",
					Usage: "Unzip the downloaded file",
				},
			},
			Action: downloadHandler,
		},
		{
			Name:    "json",
			Aliases: []string{"j"},
			Usage:   "print json doc representing latest release from github api",
			Flags:   []cli.Flag{},
			Action:  jsonHandler,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatalf("%v\n", err)
	}
}

func matchingMap(needle *regexp.Regexp, haystack string) (map[string]string, bool) {
	// generally applicable regexp utility function, takes a regexp and string
	// to search, returns a map of named capture groups
	matches := make(map[string]string)

	match := needle.FindStringSubmatch(haystack)
	if match == nil {
		return matches, false
	}

	for i, name := range needle.SubexpNames() {
		if name == "" {
			continue
		}

		matches[name] = match[i]
	}

	return matches, true
}

func httpContents(url string) (contents []byte, err error) {
	// retuns the entire contents of the given url

	resp, err := http.Get(url)
	if err != nil {
		logrus.Fatal(err)
	}
	defer resp.Body.Close()

	contents, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Fatal(err)
	}

	return contents, nil
}

func downloadFile(url string, filepath string, mode os.FileMode) (err error) {
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
	outputFH, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("Couldn't open '%s' for writing. Error: %v", filepath, err)
	}

	// Writer the body to file
	_, err = io.Copy(outputFH, resp.Body)
	if err != nil {
		return fmt.Errorf("Couldn't copy download data to output file '%v'. Error: %v", outputFH, err)
	}

	// cleanup
	outputFH.Close()

	return nil
}

func latestReleasedAssets(owner string, repo string, filter *regexp.Regexp) []string {
	// given a github owner & repo name, return a list of assets from the
	// latest release, optionally filtering results that match the given
	// filter regexp

	// logrus.SetLevel(logrus.DebugLevel)
	var result []string

	// talk to the github api and get info on the latest release
	client := github.NewClient(nil)
	ctx := context.Background()
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		logrus.Fatalf("Repositories.GetLatestRelease returned error: %v\n", err)
	}
	for _, asset := range release.Assets {
		assetName := asset.GetName()
		if filter != nil {
			if filter.MatchString(assetName) != true {
				continue
			}
		}
		result = append(result, asset.GetBrowserDownloadURL())
	}

	return result
}

func repoURLInfo(repoURL string) (owner string, repo string, err error) {
	// given a url: return the owner name, repo name, and a success indicator
	m, matched := matchingMap(repoRegexp, repoURL)
	if !matched {
		return "", "", fmt.Errorf("Invalid repo URL: '%s', it must match the regex'%s'", repoURL, repoRegexpStr)
	}

	return m[`owner`], m[`repo`], nil
}

func jsonHandler(c *cli.Context) error {
	// extract the owner and repo names from the given URL argument
	if c.NArg() != 1 {
		return fmt.Errorf("You must supply a repo URL argument")
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
		return fmt.Errorf("You must supply a repo URL argument")
	}
	owner, repo, err := repoURLInfo(c.Args().Get(0))
	if err != nil {
		return err
	}

	// process the filter argument
	var filter *regexp.Regexp
	if c.String("filter") != "" {
		filter = regexp.MustCompile(c.String("filter"))
	}
	for _, assetURL := range latestReleasedAssets(owner, repo, filter) {
		fmt.Printf("%s\n", assetURL)
	}

	return nil
}

func downloadHandler(c *cli.Context) error {
	// process the repoURL argument
	// make sure the URL looks OK and extract the owner and repo from it
	// fixme: I hate using this as a list... is there a better way?
	if c.NArg() != 1 {
		return fmt.Errorf("You must supply a repo URL argument")
	}
	owner, repo, err := repoURLInfo(c.Args().Get(0))
	if err != nil {
		return err
	}

	// process the filter argument
	var filter *regexp.Regexp
	if c.String("filter") != "" {
		filter = regexp.MustCompile(c.String("filter"))
	}

	// determine the assetsURL
	assets := latestReleasedAssets(owner, repo, filter)
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
		if filenameRegexp.MatchString(outputpath) != true {
			return fmt.Errorf("Could not correctly calculate an output filename from %s", assetURL)
		}
	}

	// process the mode argument
	// fixme: consider making this a function and adding support for symbolic modes
	mode, err := strconv.ParseUint(c.String("mode"), 8, 32)
	if err != nil {
		return fmt.Errorf("Could not process given mode string %s", c.String("mode"))
	}

	// do the download deed
	err = downloadFile(assets[0], outputpath, os.FileMode(mode))
	if err != nil {
		return err
	}
	fmt.Printf("wrote to '%s'\n", outputpath)

	return nil
}
