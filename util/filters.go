package util

import (
	"fmt"
	"regexp"
)

// FilterSet contains compiled regular expressions which identify files to be
// selected. If the FilterSet is empty all files will be selected
type FilterSet []*regexp.Regexp

// compileFilters takes a list of regular expression strings and compiles them into a [FilterSet].
func CompileFilters(filterStrings []string) (FilterSet, error) {
	filters := make(FilterSet, 0, len(filterStrings))
	for _, filterStr := range filterStrings {
		filter, err := regexp.Compile(filterStr)
		if err != nil {
			return nil, fmt.Errorf("failed to compile filter: \"%s\" - error: %s", filterStr, err)
		}
		filters = append(filters, filter)
	}
	return filters, nil
}
