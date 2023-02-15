package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
)

func latestReleasedAssets(owner string, repo string, filters []*regexp.Regexp) []string {
	// given a github owner & repo name, return a list of assets from the
	// latest release, optionally filtering results that match the given
	// filter regexp

	logrus.Debugf("Listing %s/%s with filters: %v", owner, repo, filters)
	fmt.Printf("Listing %s/%s with %d filters: %v\n", owner, repo, len(filters), filters)

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
		for _, filter := range filters {
			if !filter.MatchString(assetName) {
				goto CONTINUE_OUTER
			}
		}
		result = append(result, asset.GetBrowserDownloadURL())
	CONTINUE_OUTER:
	}

	return result
}

func repoURLInfo(repoURL string) (owner string, repo string, err error) {
	// given a url: return the owner name, repo name, and a success indicator
	m, matched := matchingMap(repoRegexp, repoURL)
	if !matched {
		return "", "", fmt.Errorf("invalid repo URL: '%s', it must match the regex'%s'", repoURL, repoRegexpStr)
	}

	return m[`owner`], m[`repo`], nil
}
