package test

import (
	"github.com/jessevdk/go-flags"
	"github.com/rumanzo/bt2qbt/internal/options"
	"reflect"
	"testing"
)

type TestArgsCase struct {
	Name     string
	Args     []string
	Opts     *options.Opts
	ErrMsg   string
	MustFail bool
	Expected *options.Opts
}

func TestOptionsArgs(t *testing.T) {
	cases := []TestArgsCase{
		{
			Name:     "Must fail test",
			Args:     []string{""},
			MustFail: true,
			Expected: &options.Opts{},
		},
		{
			Name:     "Parse without expected args test",
			Args:     []string{""},
			MustFail: false,
		},
		{
			Name: "Parse short args test",
			Args: []string{
				"-s", "/dir",
				"-d", "/dir",
				"-c", "/dir/q.conf",
				"-r", "dir1,dir2", "-r", "dir3,dir4",
				"--sep", "/",
				"-t", "/dir5", "-t", "/dir6/",
				"--without-tags"},
			MustFail: false,
			Expected: &options.Opts{
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
			Name: "Parse long args test",
			Args: []string{
				"--source", "/dir",
				"--destination", "/dir",
				"--config", "/dir/q.conf",
				"--replace", "dir1,dir2", "-r", "dir3,dir4",
				"--sep", "/",
				"--search", "/dir5", "-t", "/dir6/",
				"--without-tags"},
			MustFail: false,
			Expected: &options.Opts{
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
		opts := options.PrepareOpts()
		if _, err := flags.ParseArgs(opts, testCase.Args); err != nil { // https://godoc.org/github.com/jessevdk/go-flags#ErrorType
			if flagsErr, ok := err.(*flags.Error); !(ok && flagsErr.Type == flags.ErrHelp) && !testCase.MustFail {
				t.Fatalf("%s:\nUnexpected error: %v", testCase.Name, err)
			}
		}
		if testCase.Expected != nil {
			if !reflect.DeepEqual(testCase.Expected, opts) && !testCase.MustFail {
				t.Fatalf("%s:\nUnexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", testCase.Name, opts, testCase.Expected)
			}
		}
	}
}

func TestOptionsHandle(t *testing.T) {
	cases := []TestArgsCase{
		{
			Name: "Must fail test",
			Opts: &options.Opts{
				BitDir:        "/dir",
				QBitDir:       "/dir",
				Config:        "/dir/q.conf",
				Replaces:      []string{"dir1,dir2", "dir3,dir4"},
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/"},
				WithoutTags:   true,
			},
			MustFail: true,
			Expected: &options.Opts{},
		},
		{
			Name: "Must fail test",
			Opts: &options.Opts{
				BitDir:        "/bitdir",
				QBitDir:       "/qbitdir",
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/"},
			},
			MustFail: false,
			Expected: &options.Opts{
				BitDir:        "/bitdir",
				QBitDir:       "/qbitdir",
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/", "/bitdir"},
			},
		},
	}

	for _, testCase := range cases {
		options.HandleOpts(testCase.Opts)
		if testCase.Expected != nil {
			if !reflect.DeepEqual(testCase.Expected, testCase.Opts) && !testCase.MustFail {
				t.Fatalf("%s:\nUnexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", testCase.Name, testCase.Opts, testCase.Expected)
			}
		}
	}
}

func TestOptionsChecks(t *testing.T) {
	cases := []TestArgsCase{
		{
			Name: "Must fail don't exists folders/files test",
			Opts: &options.Opts{
				BitDir:        "/dir",
				QBitDir:       "/dir",
				Config:        "/dir/q.conf",
				Replaces:      []string{"dir1,dir2", "dir3,dir4"},
				PathSeparator: "/",
				SearchPaths:   []string{"/dir5", "/dir6/"},
				WithoutTags:   true,
			},
			MustFail: true,
		},
		{
			Name: "Check exists folders/files test",
			Opts: &options.Opts{
				BitDir:  "./data",
				QBitDir: "./data",
				Config:  "./data/testfileset.torrent",
			},
			MustFail: false,
		},
		{
			Name: "Must fail don't exists folders/files test",
			Opts: &options.Opts{
				BitDir:      "/dir",
				QBitDir:     "/dir",
				Config:      "/dir/q.conf",
				Replaces:    []string{"dir1,dir2,dir4", "dir4"},
				SearchPaths: []string{"/dir5", "/dir6/"},
			},
			MustFail: true,
		},
		{
			Name: "Must fail don't exists config test",
			Opts: &options.Opts{
				BitDir:  "./data",
				QBitDir: "./data",
				Config:  "/dir/q.conf",
			},
			MustFail: true,
		},
		{
			Name: "Must fail don't exists qbitdir test",
			Opts: &options.Opts{
				BitDir:      "./data",
				QBitDir:     "/dir",
				WithoutTags: true,
			},
			MustFail: false,
		},
	}

	for _, testCase := range cases {
		err := options.OptsCheck(testCase.Opts)
		if err != nil && !testCase.MustFail {
			t.Errorf("%s:\nUnexpected error: %v\n", testCase.Name, err)
		}
	}
}
