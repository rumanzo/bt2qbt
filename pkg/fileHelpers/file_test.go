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
			name:      "009 windows relative paths join. check normalize",
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
