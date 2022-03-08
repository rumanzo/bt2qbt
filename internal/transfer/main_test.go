package transfer

import (
	"github.com/rumanzo/bt2qbt/internal/libtorrent"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"reflect"
	"testing"
)

func TestSearchPaths(t *testing.T) {
	type SearchPathCase struct {
		name                string
		mustFail            bool
		newTorrentStructure libtorrent.NewTorrentStructure
		SearchPaths         []string
	}
	cases := []SearchPathCase{
		{
			name: "Find relative torrent directly",
			newTorrentStructure: libtorrent.NewTorrentStructure{
				TorrentFilePath: "../../test/data/testfileset.torrent",
			},
		},
		{
			name:     "Find relative torrent directly. mustFail",
			mustFail: true,
			newTorrentStructure: libtorrent.NewTorrentStructure{
				TorrentFilePath: "../../test/data/testfileset_not_existing.torrent",
			},
		},
		{
			name: "Find relative torrent with search paths",
			newTorrentStructure: libtorrent.NewTorrentStructure{
				TorrentFilePath: "",
				TorrentFileName: "testfileset.torrent",
			},
			SearchPaths: []string{"/not-exists", "../../test/data"},
		},
		{
			name:     "Find relative not existing torrent with search paths. mustFail",
			mustFail: true,
			newTorrentStructure: libtorrent.NewTorrentStructure{
				TorrentFilePath: "",
				TorrentFileName: "testfileset_not_existing.torrent",
			},
			SearchPaths: []string{"/not-exists", "../../test/data"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := FindTorrentFile(&testCase.newTorrentStructure, testCase.SearchPaths); err != nil && !testCase.mustFail {
				t.Fatalf("Unexpected error: %v", err)
			} else if testCase.mustFail && err == nil {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}

func TestHandleTorrentFilePath(t *testing.T) {
	type SearchPathCase struct {
		name                string
		mustFail            bool
		newTorrentStructure *libtorrent.NewTorrentStructure
		key                 string
		opts                *options.Opts
		expected            *libtorrent.NewTorrentStructure
	}

	cases := []SearchPathCase{
		{
			name:                "Check absolute windows path with two start backslash",
			key:                 "C:\\\\temp\\t.torrent",
			newTorrentStructure: &libtorrent.NewTorrentStructure{},
			expected: &libtorrent.NewTorrentStructure{
				TorrentFilePath: "C:\\\\temp\\t.torrent",
				TorrentFileName: "t.torrent",
			},
			opts: &options.Opts{},
		},
		{
			name:                "Check absolute windows path with two start backslash. Mustfail",
			key:                 "C:\\\\temp\\t.torrent",
			mustFail:            true,
			newTorrentStructure: &libtorrent.NewTorrentStructure{},
			expected: &libtorrent.NewTorrentStructure{
				TorrentFilePath: "C:\\temp\\t.torrent",
				TorrentFileName: "t.torrent",
			},
			opts: &options.Opts{},
		},
		{
			name:                "Check absolute windows path with one start backslash",
			key:                 "C:\\temp\\t.torrent",
			newTorrentStructure: &libtorrent.NewTorrentStructure{},
			expected: &libtorrent.NewTorrentStructure{
				TorrentFilePath: "C:\\temp\\t.torrent",
				TorrentFileName: "t.torrent",
			},
			opts: &options.Opts{},
		},
		{
			name:                "Check absolute windows path with slashes",
			key:                 "C:/temp/t.torrent",
			newTorrentStructure: &libtorrent.NewTorrentStructure{},
			expected: &libtorrent.NewTorrentStructure{
				TorrentFilePath: "C:/temp/t.torrent",
				TorrentFileName: "t.torrent",
			},
			opts: &options.Opts{},
		},
		{
			name:                "Check absolute windows share path",
			key:                 "\\temp\\t.torrent",
			newTorrentStructure: &libtorrent.NewTorrentStructure{},
			expected: &libtorrent.NewTorrentStructure{
				TorrentFilePath: "\\temp\\t.torrent",
				TorrentFileName: "t.torrent",
			},
			opts: &options.Opts{},
		},
		{
			name:                "Check relative paths (torrent name)",
			key:                 "t.torrent",
			newTorrentStructure: &libtorrent.NewTorrentStructure{},
			expected: &libtorrent.NewTorrentStructure{
				TorrentFilePath: "C:\\temp\\t.torrent",
				TorrentFileName: "t.torrent",
			},
			opts: &options.Opts{BitDir: "C:\\temp"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			HandleTorrentFilePath(testCase.newTorrentStructure, testCase.key, testCase.opts)
			equal := reflect.DeepEqual(testCase.expected, testCase.newTorrentStructure)
			if !equal && !testCase.mustFail {
				t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", testCase.newTorrentStructure, testCase.expected)
			} else if equal && testCase.mustFail {
				t.Fatal("Unexpected error: structures are equal, but they shouldn't\n", testCase.newTorrentStructure, testCase.expected)
			}
		})
	}
}

func TestPath(t *testing.T) {
	nts := libtorrent.CreateEmptyNewTorrentStructure()
	err := helpers.DecodeTorrentFile("../../test/data/testfileset.torrent", nts)
	if err != nil {
		t.Fatalf("Can't decode torrent file with error: %v", err)
	}
}
