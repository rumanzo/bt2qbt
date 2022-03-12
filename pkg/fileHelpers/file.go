package fileHelpers

/* Default go filepath works wrong with some windows paths like windows shares (\\somepath), use only os.PathSeparator and so on*/
import (
	"path/filepath"
	"regexp"
	"strings"
)

var checkAbsRegExp = regexp.MustCompile(`^(([A-Za-z]:)(\\\\?|/)|(\\\\|//))`)

var checkIsShare = regexp.MustCompile(`^(//|\\\\)`)

func IsAbs(filePath string) bool {
	if checkAbsRegExp.MatchString(filePath) {
		return true
	}
	return false
}

func Join(filePaths []string, separator string) string {
	var filePath string
	var prefix string
	if checkIsShare.MatchString(filePaths[0]) {
		prefix = filePaths[0][:2]
		filePaths[0] = filePaths[0][2:]
		filePath = filepath.Join(filePaths...)
	}
	filePath = filepath.Join(filePaths...)
	//normalize separator
	filePath = filepath.ToSlash(filePath)
	filePath = prefix + filePath
	if separator == "/" {
		filePath = strings.ReplaceAll(filePath, `\`, `/`)
	} else {
		filePath = strings.ReplaceAll(filePath, `/`, `\`)
	}
	return filePath
}

func Base(filePath string) string {
	if checkIsShare.MatchString(filePath) {
		return filepath.Base(filePath[2:])
	}
	return filepath.Base(filePath)
}

func Normalize(filePath string, separator string) string {
	var prefix string
	if checkIsShare.MatchString(filePath) {
		prefix = filePath[:2]
		filePath = filePath[2:]
	}
	filePath = filepath.Clean(filePath)
	filePath = prefix + filePath
	if separator == "/" {
		filePath = strings.ReplaceAll(filePath, `\`, `/`)
	} else {
		filePath = strings.ReplaceAll(filePath, `/`, `\`)
	}
	return filePath
}
