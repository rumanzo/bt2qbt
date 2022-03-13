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

func Base(filePath string) string {
	if checkIsShareRegExp.MatchString(filePath) {
		filePath = filePath[2:]
	}
	return filepath.Base(filePath)
}

func Normalize(filePath string, separator string) string {
	var prefix string
	if checkIsShareRegExp.MatchString(filePath) {
		prefix = filePath[:2]
		filePath = filePath[2:]
	}
	filePath = filepath.Clean(filePath)
	filePath = prefix + filePath
	if separator == "/" {
		filePath = filepath.ToSlash(filePath)
	} else {
		filePath = strings.ReplaceAll(filePath, `/`, `\`)
	}
	return filePath
}

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
	if lastSeparatorIndex < 0 {
		return filePath
	}
	return filePath[:lastSeparatorIndex]
}
