package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/go-github/v33/github"
	log "github.com/sirupsen/logrus"
)

func latestReleasedAssets(owner string, repo string, filters []*regexp.Regexp, source bool) []string {
	// given a github owner & repo name, return a list of assets from the
	// latest release, optionally filtering results that match the given
	// filter regexp

	var result []string

	log.Debugf("Listing %s/%s with %d filters: %v", owner, repo, len(filters), filters)
	// talk to the github api and get info on the latest release
	client := github.NewClient(nil)
	ctx := context.Background()
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		log.Fatalf("Repositories.GetLatestRelease returned error: %v\n", err)
	}
	if source {
		result = append(result, release.GetTarballURL())
		return result
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
