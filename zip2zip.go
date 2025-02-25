package zip2zip

import (
	"regexp"
)

type Pattern struct {
	*regexp.Regexp
}

type FilterResult uint8

const (
	FilterResultUnspecified FilterResult = 0
	FilterResultKeep        FilterResult = 1
	FilterResultSkip        FilterResult = 2
)

var PatternDefault Pattern = Pattern{
	Regexp: regexp.MustCompile("."),
}
