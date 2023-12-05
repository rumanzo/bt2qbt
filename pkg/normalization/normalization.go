package normalization

import (
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"regexp"
)

// ProhibitedSymbolsStrict we can't use these symbols on Windows systems, but can use in *nix
var ProhibitedSymbolsStrict = regexp.MustCompilePOSIX(`[\\/:*?"<>|]`)

func NormalizeSpaceEnding(str string) (string, bool) {
	var normalized bool
	if string(str[len(str)-1]) == ` ` {
		str = str[:len(str)-1] + `_`
		normalized = true
	}
	return str, normalized
}

func FullNormalize(str string) (string, bool) {
	var normalized bool
	s1 := ProhibitedSymbolsStrict.ReplaceAllString(str, `_`)
	if s1 != str {
		normalized = true
	}
	s2 := helpers.HandleCesu8(s1)
	if s1 != s2 {
		normalized = true
	}
	s3, n := NormalizeSpaceEnding(s2)
	if n {
		normalized = true
	}
	return s3, normalized
}
