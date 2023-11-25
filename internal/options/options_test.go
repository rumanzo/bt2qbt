package options

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/jessevdk/go-flags"
	"github.com/r3labs/diff/v2"
	"reflect"
	"testing"
)

type TestArgsCase struct {
	name     string
	args     []string
	opts     *Opts
	errMsg   string
	mustFail bool
	expected *Opts
}

func TestOptionsArgs(t *testing.T) {
	cases := []TestArgsCase{
		{
			name:     "Must fail test",
			args:     []string{""},
			mustFail: true,
			expected: &Opts{},
		},
		{
			name:     "Parse without expected args test",
			args:     []string{""},
			mustFail: false,
		},
		{
			name: "Parse short args test",
			args: []string{
				"-s", "/dir",
				"-d", "/dir",
				"-c", "/dir/q.json",
				"-r", "dir1,dir2", "-r", "dir3,dir4",
				"--sep", "/",
				"-t", "/dir5", "-t", "/dir6/",
				"--without-tags"},
			mustFail: false,
			expected: &Opts{
				BitDir:        "/dir",
				QBitDir:       "/dir",
				Categories:    "/dir/q.json",
				Replaces:      []string{"dir1,dir2", "dir3,dir4"},
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/"},
				WithoutTags:   true,
			},
		},
		{
			name: "Parse long args test",
			args: []string{
				"--source", "/dir",
				"--destination", "/dir",
				"--categories", "/dir/q.json",
				"--replace", "dir1,dir2", "-r", "dir3,dir4",
				"--sep", "/",
				"--search", "/dir5", "-t", "/dir6/",
				"--without-tags"},
			mustFail: false,
			expected: &Opts{
				BitDir:        "/dir",
				QBitDir:       "/dir",
				Categories:    "/dir/q.json",
				Replaces:      []string{"dir1,dir2", "dir3,dir4"},
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/"},
				WithoutTags:   true,
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			opts := PrepareOpts()
			if _, err := flags.ParseArgs(opts, testCase.args); err != nil { // https://godoc.org/github.com/jessevdk/go-flags#ErrorType
				if flagsErr, ok := err.(*flags.Error); !(ok && flagsErr.Type == flags.ErrHelp) && !testCase.mustFail {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
			if testCase.expected != nil {
				if !reflect.DeepEqual(testCase.expected, opts) && !testCase.mustFail {
					changes, err := diff.Diff(opts, testCase.expected, diff.DiscardComplexOrigin())
					if err != nil {
						t.Error(err.Error())
					}
					t.Fatalf("Unexpected error: opts isn't equoptions:\nGot: %#v\nExpect %#v\nDiff: %v\\n", opts, testCase.expected, spew.Sdump(changes))
				}
			}
		})
	}
}

func TestOptionsHandle(t *testing.T) {
	cases := []TestArgsCase{
		{
			name: "001 Must fail test",
			opts: &Opts{
				BitDir:        "/dir",
				QBitDir:       "/dir",
				Categories:    "/dir/q.json",
				Replaces:      []string{"dir1,dir2", "dir3,dir4"},
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/"},
				WithoutTags:   true,
			},
			mustFail: true,
			expected: &Opts{},
		},
		{
			name: "002 Must fail test",
			opts: &Opts{
				BitDir:        "/bitdir",
				QBitDir:       "/qbitdir",
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/"},
			},
			mustFail: true,
			expected: &Opts{
				BitDir:        "/bitdir",
				QBitDir:       "/qbitdir",
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/", "/bitdir"},
			},
		},
		{
			name: "003 Parse portable args test",
			opts: &Opts{
				BitDir:        `/dir1`,
				QBitDir:       `C:\btportable\profile\qBittorrent\data\BT_backup\`,
				PathSeparator: `\`,
				SearchPaths:   []string{},
			},
			mustFail: false,
			expected: &Opts{
				BitDir:        `/dir1`,
				QBitDir:       `C:\btportable\profile\qBittorrent\data\BT_backup\`,
				Categories:    `C:\btportable\profile\qBittorrent\config\categories.json`,
				SearchPaths:   []string{`/dir1`},
				PathSeparator: `\`,
			},
		},
		{
			name: "004 Parse portable args test with categories file",
			opts: &Opts{
				BitDir:        `/dir1`,
				QBitDir:       `C:\btportable\profile\qBittorrent\data\BT_backup\`,
				Categories:    `C:\categories.json`,
				PathSeparator: `\`,
				SearchPaths:   []string{},
			},
			mustFail: false,
			expected: &Opts{
				BitDir:        `/dir1`,
				QBitDir:       `C:\btportable\profile\qBittorrent\data\BT_backup\`,
				Categories:    `C:\categories.json`,
				SearchPaths:   []string{`/dir1`},
				PathSeparator: `\`,
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.opts.Categories == `` {
				refOpts := PrepareOpts()
				testCase.opts.Categories = refOpts.Categories
			}
			HandleOpts(testCase.opts)
			if testCase.expected != nil {
				changes, err := diff.Diff(testCase.opts, testCase.expected, diff.DiscardComplexOrigin())
				if err != nil {
					t.Error(err.Error())
				}
				if !reflect.DeepEqual(testCase.expected, testCase.opts) && !testCase.mustFail {
					t.Fatalf("Unexpected error: opts isn't equal:\nGot: %#v\nExpect %#v\nDiff: %v\\\\n", testCase.opts, testCase.expected, spew.Sdump(changes))
				}
			}
		})
	}
}

func TestOptionsChecks(t *testing.T) {
	cases := []TestArgsCase{
		{
			name: "001 Must fail don't exists folders or files",
			opts: &Opts{
				BitDir:        "/dir",
				QBitDir:       "/dir",
				Categories:    "/dir/q.json",
				Replaces:      []string{"dir1,dir2", "dir3,dir4"},
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/"},
				WithoutTags:   true,
			},
			mustFail: true,
		},
		{
			name: "002 Check exists folders or files",
			opts: &Opts{
				BitDir:      "../../test/data",
				QBitDir:     "../../test/data",
				SearchPaths: []string{},
			},
			mustFail: false,
		},
		{
			name: "003 Must fail do not exists folders or files test",
			opts: &Opts{
				BitDir:      "/dir",
				QBitDir:     "/dir",
				Categories:  "/dir/q.json",
				Replaces:    []string{"dir1,dir2,dir4", "dir4"},
				SearchPaths: []string{"/dir5", "/dir6/"},
			},
			mustFail: true,
		},
		{
			name: "004 Must fail do not exists qbitdir test",
			opts: &Opts{
				BitDir:      "../../test/data",
				QBitDir:     "/dir",
				WithoutTags: true,
			},
			mustFail: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := OptsCheck(testCase.opts)
			if err != nil && !testCase.mustFail {
				t.Errorf("Unexpected error: %v\n", err)
			}
		})
	}
}
