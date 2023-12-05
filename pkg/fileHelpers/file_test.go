package fileHelpers

import (
	"testing"
)

func TestIsAbs(t *testing.T) {
	type PathAbsCase struct {
		name     string
		mustFail bool
		path     string
		expected bool
	}
	cases := []PathAbsCase{
		{
			name:     "001 Full windows backslash path with file with spaces",
			path:     `C:\\testdir\my test file.txt`,
			expected: true,
		},
		{
			name:     "002 Full windows backslash path with file with spaces. Mustfail",
			mustFail: true,
			path:     `C:\\testdir\my test file.txt`,
			expected: false,
		},
		{
			name:     "003 Full windows share backslash path with file with spaces",
			path:     `\\testdir\my test file.txt`,
			expected: true,
		},
		{
			name:     "004 Full windows slash path with file with spaces",
			path:     `C:/testdir/my test file.txt`,
			expected: true,
		},
		{
			name:     "005 Full short slash path with file with spaces",
			path:     `C:/my test file.txt`,
			expected: true,
		},
		{
			name:     "006 Relative slash path with file with spaces",
			path:     `./testdir/my test file.txt`,
			expected: false,
		},
		{
			name:     "007 Relative backslash path with file with spaces",
			path:     `.\testdir\my test file.txt`,
			expected: false,
		},
		{
			name:     "008 Relative wrong backslash path with file with spaces",
			path:     `.\\testdir\my test file.txt`,
			expected: false,
		},
		{
			name:     "009 Full linux windows share backslash path with file with spaces",
			path:     `//testdir/my test file.txt`,
			expected: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if isAbs := IsAbs(testCase.path); isAbs != testCase.expected && !testCase.mustFail {
				t.Fatalf("Unexpected error: should be %v, got %v", testCase.expected, isAbs)
			} else if testCase.mustFail && isAbs == testCase.expected {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}

func TestJoin(t *testing.T) {
	type PathJoinCase struct {
		name      string
		mustFail  bool
		paths     []string
		separator string
		expected  string
	}
	cases := []PathJoinCase{
		{
			name:      "001 windows paths join",
			paths:     []string{`C:\\`, `testdir\my test file.txt`},
			separator: `\`,
			expected:  `C:\testdir\my test file.txt`,
		},
		{
			name:      "002 windows paths join. Mustfail",
			paths:     []string{`C:\\`, `testdir\my test file.txt`},
			separator: `\`,
			expected:  `C:\testdir\\my test file.txt`,
			mustFail:  true,
		},
		{
			name:      "003 windows share paths join",
			paths:     []string{`\\`, `testdir\my test file.txt`},
			separator: `\`,
			expected:  `\\testdir\my test file.txt`,
		},
		{
			name:      "004 linux share paths join",
			paths:     []string{`//`, `testdir/my test file.txt`},
			separator: `/`,
			expected:  `//testdir/my test file.txt`,
		},
		{
			name:      "005 linux share paths join",
			paths:     []string{`//testdir`, `/my test file.txt`},
			separator: `/`,
			expected:  `//testdir/my test file.txt`,
		},
		{
			name:      "006 linux share paths join",
			paths:     []string{`//testdir`, `my test file.txt`},
			separator: `/`,
			expected:  `//testdir/my test file.txt`,
		},
		{
			name:      "007 windows relative paths join revert slash",
			paths:     []string{`../../testdir`, `my test file.txt`},
			separator: `\`,
			expected:  `..\..\testdir\my test file.txt`,
		},
		{
			name:      "008 windows relative paths join",
			paths:     []string{`..\..\testdir`, `my test file.txt`},
			separator: `\`,
			expected:  `..\..\testdir\my test file.txt`,
		},
		{
			name:      "009 linux relative paths join. check normalize",
			paths:     []string{`./testdir/../testdir/../testdir/./`, `my test file.txt`},
			separator: `/`,
			expected:  `testdir/my test file.txt`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if joined := Join(testCase.paths, testCase.separator); testCase.expected != joined && !testCase.mustFail {
				t.Fatalf("Unexpected error: should be %v, got %v", testCase.expected, joined)
			} else if testCase.mustFail && joined == testCase.expected {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}
func TestBase(t *testing.T) {
	type PathBaseCase struct {
		name     string
		mustFail bool
		path     string
		expected string
	}
	cases := []PathBaseCase{
		{
			name:     "001 empty path",
			path:     ``,
			expected: `.`,
		},
		{
			name:     "002 empty path. Mustfail",
			path:     ``,
			mustFail: true,
			expected: ``,
		},
		{
			name:     "003 Full windows path with file ending with backslash",
			path:     `C:\\mydir\myfile.txt`,
			expected: `myfile.txt`,
		},
		{
			name:     "004 Full windows path with backslash",
			path:     `C:\\mydir`,
			expected: `mydir`,
		},
		{
			name:     "005 Short windows path with backslash",
			path:     `C:\mydir\myfile.txt`,
			expected: `myfile.txt`,
		},
		{
			name:     "006 Short windows path with backslash",
			path:     `C:\mydir\myfile.txt`,
			expected: `myfile.txt`,
		},
		{
			name:     "007 Short windows path with slash",
			path:     `C:/mydir/myfile.txt`,
			expected: `myfile.txt`,
		},
		{
			name:     "008 Windows share path with slashes",
			path:     `//mydir/myfile.txt`,
			expected: `myfile.txt`,
		},
		{
			name:     "009 Windows share path with backslashes",
			path:     `\\mydir\myfile.txt`,
			expected: `myfile.txt`,
		},
		{
			name:     "010 Windows share path with backslashes directory",
			path:     `\\mydir\\mydir2\\`,
			expected: `mydir2`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if base := Base(testCase.path); testCase.expected != base && !testCase.mustFail {
				t.Fatalf("Unexpected error: should be %#v, got %#v", testCase.expected, base)
			} else if testCase.mustFail && base == testCase.expected {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}
func TestNormalize(t *testing.T) {
	type PathNormalizeCase struct {
		name      string
		mustFail  bool
		separator string
		path      string
		expected  string
	}
	cases := []PathNormalizeCase{
		{
			name:      "001 empty path",
			path:      ``,
			separator: "/",
			expected:  `.`,
		},
		{
			name:      "002 empty path. Mustfail",
			path:      ``,
			separator: "/",
			mustFail:  true,
			expected:  ``,
		},
		{
			name:      "003 Full windows path with file ending with backslash",
			separator: `\`,
			path:      `C:\\mydir\myfile.txt`,
			expected:  `C:\mydir\myfile.txt`,
		},
		{
			name:      "004 Full windows path with backslash",
			separator: `\`,
			path:      `C:\\mydir`,
			expected:  `C:\mydir`,
		},
		{
			name:      "005 Short windows path with backslash",
			separator: `\`,
			path:      `C:\mydir\myfile.txt`,
			expected:  `C:\mydir\myfile.txt`,
		},
		{
			name:      "006 Short windows path with backslash and backslash ending",
			separator: `\`,
			path:      `C:\mydir\`,
			expected:  `C:\mydir`,
		},
		{
			name:      "007 Short windows path with slash",
			separator: `\`,
			path:      `C:/mydir/myfile.txt`,
			expected:  `C:\mydir\myfile.txt`,
		},
		{
			name:      "008 Windows share path with slashes. Change separator",
			separator: `\`,
			path:      `//mydir/myfile.txt`,
			expected:  `\\mydir\myfile.txt`,
		},
		{
			name:      "009 Windows share path with backslashes. Change separator",
			separator: `/`,
			path:      `\\mydir\myfile.txt`,
			expected:  `//mydir/myfile.txt`,
		},
		{
			name:      "010 Windows share path with slashes.",
			separator: `/`,
			path:      `//mydir/myfile.txt`,
			expected:  `//mydir/myfile.txt`,
		},
		{
			name:      "011 Windows share path with backslashes",
			separator: `\`,
			path:      `\\mydir\myfile.txt`,
			expected:  `\\mydir\myfile.txt`,
		},
		{
			name:      "012 Windows share path with backslashes directory",
			separator: `\`,
			path:      `\\mydir\\mydir2\\`,
			expected:  `\\mydir\mydir2`,
		},
		{
			name:      "013 Windows share path with backslashes directory. Change separator",
			separator: `/`,
			path:      `\\mydir\mydir2\`,
			expected:  `//mydir/mydir2`,
		},
		{
			name:      "014 Windows share path with broken backslashes directory. Change separator",
			separator: `/`,
			path:      `\\mydir\\mydir2\\`,
			expected:  `//mydir/mydir2`,
		},
		{
			name:      "015 Windows share path with slashes directory.",
			separator: `/`,
			path:      `//mydir/mydir2/`,
			expected:  `//mydir/mydir2`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if normalized := Normalize(testCase.path, testCase.separator); testCase.expected != normalized && !testCase.mustFail {
				t.Fatalf("Unexpected error: should be %#v, got %#v", testCase.expected, normalized)
			} else if testCase.mustFail && normalized == testCase.expected {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}
func TestCutLastPath(t *testing.T) {
	type PathCutLastPathCase struct {
		name      string
		mustFail  bool
		separator string
		path      string
		expected  string
	}
	cases := []PathCutLastPathCase{
		{
			name:      "001 empty path",
			path:      ``,
			separator: "/",
			expected:  ``,
		},
		{
			name:      "002 empty path. Mustfail",
			path:      ``,
			separator: "/",
			mustFail:  true,
			expected:  `.`,
		},
		{
			name:      "003 Full windows path with file ending with backslash",
			separator: `\`,
			path:      `C:\\mydir\myfile.txt`,
			expected:  `C:\mydir`,
		},
		{
			name:      "004 Full windows path with backslash",
			separator: `\`,
			path:      `C:\\mydir`,
			expected:  `C:\`,
		},
		{
			name:      "005 Short windows path with backslash",
			separator: `\`,
			path:      `C:\mydir\myfile.txt`,
			expected:  `C:\mydir`,
		},
		{
			name:      "006 Short windows path with backslash and backslash ending",
			separator: `\`,
			path:      `C:\mydir\`,
			expected:  `C:\`,
		},
		{
			name:      "007 Short windows path with slash",
			separator: `\`,
			path:      `C:/mydir/myfile.txt`,
			expected:  `C:\mydir`,
		},
		{
			name:      "008 Windows share path with slashes. Change separator",
			separator: `\`,
			path:      `//mydir/myfile.txt`,
			expected:  `\\mydir`,
		},
		{
			name:      "009 Windows share path with backslashes. Change separator",
			separator: `/`,
			path:      `\\mydir\myfile.txt`,
			expected:  `//mydir`,
		},
		{
			name:      "010 Windows share path with slashes.",
			separator: `/`,
			path:      `//mydir/myfile.txt`,
			expected:  `//mydir`,
		},
		{
			name:      "011 Windows share path with backslashes",
			separator: `\`,
			path:      `\\mydir\myfile.txt`,
			expected:  `\\mydir`,
		},
		{
			name:      "012 Windows share path with backslashes directory",
			separator: `\`,
			path:      `\\mydir\\mydir2\\`,
			expected:  `\\mydir`,
		},
		{
			name:      "013 Windows share path with backslashes directory. Change separator",
			separator: `/`,
			path:      `\\mydir\mydir2\`,
			expected:  `//mydir`,
		},
		{
			name:      "014 Windows share path with broken backslashes directory. Change separator",
			separator: `/`,
			path:      `\\mydir\\mydir2\\`,
			expected:  `//mydir`,
		},
		{
			name:      "015 Windows share path with slashes directory.",
			separator: `/`,
			path:      `//mydir/mydir2/`,
			expected:  `//mydir`,
		},
		{
			name:      "016 Linux path with slashes directory.",
			separator: `/`,
			path:      `/mydir`,
			expected:  `/`,
		},
		{
			name:      "017 Windows share path with broken backslashes directory.",
			separator: `\`,
			path:      `\\mydir`,
			expected:  `\\`,
		},
		{
			name:      "018 Windows vanilla path",
			separator: `/`,
			path:      `C:/`,
			expected:  `C:/`,
		},
		{
			name:      "019 Windows relative path with backslashes with split",
			separator: `\`,
			path:      `.\somepath`,
			expected:  `.\`,
		},
		{
			name:      "020 Windows relative path with backslashes without split",
			separator: `\`,
			path:      `.\`,
			expected:  `.\`,
		},
		{
			name:      "021 Linux relative path with lashes with split",
			separator: `/`,
			path:      `./somepath`,
			expected:  `./`,
		},
		{
			name:      "022 Linux relative path with slashes without split",
			separator: `/`,
			path:      `./`,
			expected:  `./`,
		},
		{
			name:      "023 Linux absolute root path with slashes without any split",
			separator: `/`,
			path:      `/`,
			expected:  `/`,
		},
		{
			name:      "023 Linux absolute root path with slashes without any split. with change separator",
			separator: `\`,
			path:      `/`,
			expected:  `\`,
		},
		{
			name:      "024 just file without any slashes or backslashes",
			separator: `/`,
			path:      `myfile.txt`,
			expected:  ``,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if cutted := CutLastPath(testCase.path, testCase.separator); testCase.expected != cutted && !testCase.mustFail {
				t.Fatalf("Unexpected error: should be %#v, got %#v", testCase.expected, cutted)
			} else if testCase.mustFail && cutted == testCase.expected {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}
