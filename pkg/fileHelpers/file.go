package fileHelpers

/* Default go filepath works wrong with some windows paths like windows shares (\\somepath), use only os.PathSeparator and so on*/
import (
	"path/filepath"
	"regexp"
	"strings"
)

var checkAbsRegExp = regexp.MustCompile(`^(([A-Za-z]:)(\\\\?|/)|(\\\\|//))`)

var checkIsShareRegExp = regexp.MustCompile(`^(//|\\\\)`)

var rootPathRegexp = regexp.MustCompile(`^(\.\.?(/|\\)|[A-Za-z]:(/|\\)|(//?|\\\\))`)

func IsAbs(filePath string) bool {
	if checkAbsRegExp.MatchString(filePath) {
		return true
	}
	return false
}

func Join(filePaths []string, separator string) string {
	var filePath string
	if checkIsShareRegExp.MatchString(filePaths[0]) {
		prefix := filePaths[0][:2]
		filePaths[0] = filePaths[0][2:]
		filePath = filepath.Join(filePaths...)
		filePath = prefix + filePath
	} else {
		filePath = filepath.Join(filePaths...)
	}
	filePath = Normalize(filePath, separator)
	return filePath
}

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns ".".
// If the path consists entirely of separators, Base returns a single separator.
func Base(filePath string) string {
	if checkIsShareRegExp.MatchString(filePath) {
		filePath = filePath[2:]
	}
	if filePath == "" {
		return "."
	}
	// Strip trailing slashes.
	for len(filePath) > 0 && IsPathSeparator(filePath[len(filePath)-1]) {
		filePath = filePath[0 : len(filePath)-1]
	}
	// Throw away volume name
	filePath = filePath[len(filepath.VolumeName(filePath)):]
	// Find the last element
	i := len(filePath) - 1
	for i >= 0 && !IsPathSeparator(filePath[i]) {
		i--
	}
	if i >= 0 {
		filePath = filePath[i+1:]
	}
	// If empty now, it had only slashes.
	if filePath == "" {
		return string(filePath)
	}
	return filePath
}

func Normalize(filePath string, separator string) string {
	var prefix string
	if checkIsShareRegExp.MatchString(filePath) {
		prefix = filePath[:2]
		filePath = filePath[2:]
	}

	filePath = strings.ReplaceAll(filePath, `\`, `/`)
	filePath = filepath.Clean(filePath)
	filePath = prefix + filePath
	if separator == `\` {
		filePath = strings.ReplaceAll(filePath, `/`, `\`)
	} else {
		// we need change separators in prefix too
		filePath = strings.ReplaceAll(filePath, `\`, `/`)
	}
	return filePath
}

// remove last dir or file, change separator, normalize and leave root path exists
func CutLastPath(filePath string, separator string) string {
	prefixSubmatch := rootPathRegexp.FindAllString(filePath, 1)
	var prefix string
	if len(prefixSubmatch) > 0 {
		prefix = prefixSubmatch[0]
		if separator == "/" {
			prefix = strings.ReplaceAll(prefix, `\`, `/`)
		} else {
			prefix = strings.ReplaceAll(prefix, `/`, `\`)
		}
	}
	filePath = Normalize(filePath, separator)
	lastSeparatorIndex := strings.LastIndex(filePath, separator)
	if lastSeparatorIndex < len(prefix) {
		return prefix
	}
	return filePath[:lastSeparatorIndex]
}

// windows reazilation everywhere
func IsPathSeparator(c uint8) bool {
	// NOTE: Windows accept / as path separator.
	return c == '\\' || c == '/'
}
