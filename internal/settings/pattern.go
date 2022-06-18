package settings

import "regexp"

type PatternConfig struct {
	RegExp   *regexp.Regexp
	Template string
}
