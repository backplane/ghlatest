package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"

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
	archRegexp     *regexp.Regexp
	osRegexp       *regexp.Regexp
)

func init() {
	var arch_subregex, os_subregex string

	// to see the available GOARCH and GOOS options, run "go tool dist list"

	switch runtime.GOARCH {
	case `amd64`:
		arch_subregex = `(amd64|x86_64)`
	case `arm64`:
		arch_subregex = `(arm64|aarch64)`
	case `arm`:
		arch_subregex = `arm(v[\d\w]{2,3})?`
	default:
		arch_subregex = runtime.GOARCH
	}
	archRegexp = regexp.MustCompile(`(?i)[^0-9a-fA-F]` + arch_subregex + `[^0-9a-fA-F]`)

	switch runtime.GOOS {
	case `darwin`:
		os_subregex = `(darwin|macos|osx)`
	case `freebsd`:
		os_subregex = `(freebsd|fbsd)`
	case `windows`:
		os_subregex = `(windows|win)`
	default:
		os_subregex = runtime.GOOS
	}
	osRegexp = regexp.MustCompile(`(?i)[^0-9a-fA-F]` + os_subregex + `[^0-9a-fA-F]`)

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
					Value: "",
					Usage: "Filter release assets with the given regular expression",
				},
				cli.StringFlag{
					Name:  "ifilter",
					Value: "",
					Usage: "Filter release assets with the given CASE-INSENSITIVE regular expression",
				},
				cli.BoolFlag{
					Name:  "current-arch",
					Usage: "Filter release assets with a regex describing the current processor architecture",
				},
				cli.BoolFlag{
					Name:  "current-os",
					Usage: "Filter release assets with a regex describing the current operating system",
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
					Value: "",
					Usage: "Filter release assets with the given regular expression",
				},
				cli.StringFlag{
					Name:  "ifilter",
					Value: "",
					Usage: "Filter release assets with the given CASE INSENSITIVE regular expression",
				},
				cli.BoolFlag{
					Name:  "current-arch",
					Usage: "Filter release assets with a regex describing the current processor architecture",
				},
				cli.BoolFlag{
					Name:  "current-os",
					Usage: "Filter release assets with a regex describing the current operating system",
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
