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
		t.Run(testCase.Name, func(t *testing.T) {
			opts := options.PrepareOpts()
			if _, err := flags.ParseArgs(opts, testCase.Args); err != nil { // https://godoc.org/github.com/jessevdk/go-flags#ErrorType
				if flagsErr, ok := err.(*flags.Error); !(ok && flagsErr.Type == flags.ErrHelp) && !testCase.MustFail {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
			if testCase.Expected != nil {
				if !reflect.DeepEqual(testCase.Expected, opts) && !testCase.MustFail {
					t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", opts, testCase.Expected)
				}
			}
		})
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
		t.Run(testCase.Name, func(t *testing.T) {
			options.HandleOpts(testCase.Opts)
			if testCase.Expected != nil {
				if !reflect.DeepEqual(testCase.Expected, testCase.Opts) && !testCase.MustFail {
					t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", testCase.Opts, testCase.Expected)
				}
			}
		})
	}
}

func TestOptionsChecks(t *testing.T) {
	cases := []TestArgsCase{
		{
			Name: "Must fail don't exists folders or files",
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
			Name: "Check exists folders or files",
			Opts: &options.Opts{
				BitDir:  "./data",
				QBitDir: "./data",
				Config:  "./data/testfileset.torrent",
			},
			MustFail: false,
		},
		{
			Name: "Must fail do not exists folders or files test",
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
			Name: "Must fail do not exists config test",
			Opts: &options.Opts{
				BitDir:  "./data",
				QBitDir: "./data",
				Config:  "/dir/q.conf",
			},
			MustFail: true,
		},
		{
			Name: "Must fail do not exists qbitdir test",
			Opts: &options.Opts{
				BitDir:      "./data",
				QBitDir:     "/dir",
				WithoutTags: true,
			},
			MustFail: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			err := options.OptsCheck(testCase.Opts)
			if err != nil && !testCase.MustFail {
				t.Errorf("Unexpected error: %v\n", err)
			}
		})
	}
}
