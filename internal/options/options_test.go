package options

import (
	"github.com/jessevdk/go-flags"
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
				"-c", "/dir/q.conf",
				"-r", "dir1,dir2", "-r", "dir3,dir4",
				"--sep", "/",
				"-t", "/dir5", "-t", "/dir6/",
				"--without-tags"},
			mustFail: false,
			expected: &Opts{
				BitDir:        "/dir",
				QBitDir:       "/dir",
				Config:        "/dir/q.conf",
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
				"--config", "/dir/q.conf",
				"--replace", "dir1,dir2", "-r", "dir3,dir4",
				"--sep", "/",
				"--search", "/dir5", "-t", "/dir6/",
				"--without-tags"},
			mustFail: false,
			expected: &Opts{
				BitDir:        "/dir",
				QBitDir:       "/dir",
				Config:        "/dir/q.conf",
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
					t.Fatalf("Unexpected error: opts isn't equoptions:\n Got: %#v\n Expect %#v\n", opts, testCase.expected)
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
				Config:        "/dir/q.conf",
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
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			HandleOpts(testCase.opts)
			if testCase.expected != nil {
				if !reflect.DeepEqual(testCase.expected, testCase.opts) && !testCase.mustFail {
					t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", testCase.opts, testCase.expected)
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
				Config:        "/dir/q.conf",
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
				Config:      "../../test/data/testfileset.torrent",
				SearchPaths: []string{},
			},
			mustFail: false,
		},
		{
			name: "003 Must fail do not exists folders or files test",
			opts: &Opts{
				BitDir:      "/dir",
				QBitDir:     "/dir",
				Config:      "/dir/q.conf",
				Replaces:    []string{"dir1,dir2,dir4", "dir4"},
				SearchPaths: []string{"/dir5", "/dir6/"},
			},
			mustFail: true,
		},
		{
			name: "004 Must fail do not exists config test",
			opts: &Opts{
				BitDir:  "../../test/data",
				QBitDir: "../../test/data",
				Config:  "/dir/q.conf",
			},
			mustFail: true,
		},
		{
			name: "005 Must fail do not exists qbitdir test",
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
