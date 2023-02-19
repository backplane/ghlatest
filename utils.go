package main

import (
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/sirupsen/logrus"
)

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
