package transfer

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/r3labs/diff/v2"
	"github.com/rumanzo/bt2qbt/internal/libtorrent"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"reflect"
	"testing"
)

func TestSearchPaths(t *testing.T) {
	type SearchPathCase struct {
		name                 string
		mustFail             bool
		newTransferStructure libtorrent.TransferStructure
		SearchPaths          []string
	}
	cases := []SearchPathCase{
		{
			name: "001 Find relative torrent directly",
			newTransferStructure: libtorrent.TransferStructure{
				TorrentFilePath: "../../test/data/testfileset.torrent",
				Opts:            &options.Opts{},
			},
		},
		{
			name:     "002 Find relative torrent directly. mustFail",
			mustFail: true,
			newTransferStructure: libtorrent.TransferStructure{
				TorrentFilePath: "../../test/data/testfileset_not_existing.torrent",
				Opts:            &options.Opts{},
			},
		},
		{
			name: "003 Find relative torrent with search paths",
			newTransferStructure: libtorrent.TransferStructure{
				TorrentFilePath: "",
				TorrentFileName: "testfileset.torrent",
				Opts:            &options.Opts{SearchPaths: []string{"/not-exists", "../../test/data"}},
			},
		},
		{
			name:     "004 Find relative not existing torrent with search paths. mustFail",
			mustFail: true,
			newTransferStructure: libtorrent.TransferStructure{
				TorrentFilePath: "",
				TorrentFileName: "testfileset_not_existing.torrent",
				Opts:            &options.Opts{SearchPaths: []string{"/not-exists", "../../test/data"}},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := FindTorrentFile(&testCase.newTransferStructure); err != nil && !testCase.mustFail {
				t.Fatalf("Unexpected error: %v", err)
			} else if testCase.mustFail && err == nil {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}

func TestHandleTorrentFilePath(t *testing.T) {
	type SearchPathCase struct {
		name                 string
		mustFail             bool
		newTransferStructure *libtorrent.TransferStructure
		key                  string
		opts                 *options.Opts
		expected             *libtorrent.TransferStructure
	}

	cases := []SearchPathCase{
		{
			name:                 "001 Check absolute windows path with two start backslash",
			key:                  `C:\\temp\t.torrent`,
			newTransferStructure: &libtorrent.TransferStructure{Opts: &options.Opts{}},
			expected: &libtorrent.TransferStructure{
				TorrentFilePath: `C:\\temp\t.torrent`,
				TorrentFileName: "t.torrent",
				Opts:            &options.Opts{},
			},
		},
		{
			name:                 "002 Check absolute windows path with two start backslash. Mustfail",
			key:                  `C:\\temp\t.torrent`,
			mustFail:             true,
			newTransferStructure: &libtorrent.TransferStructure{Opts: &options.Opts{}},
			expected: &libtorrent.TransferStructure{
				TorrentFilePath: `C:\\temp\\t.torrent`,
				TorrentFileName: "t.torrent",
				Opts:            &options.Opts{},
			},
		},
		{
			name:                 "003 Check absolute windows path with one start backslash",
			key:                  `C:\\temp\t.torrent`,
			newTransferStructure: &libtorrent.TransferStructure{Opts: &options.Opts{}},
			expected: &libtorrent.TransferStructure{
				TorrentFilePath: `C:\\temp\t.torrent`,
				TorrentFileName: "t.torrent",
				Opts:            &options.Opts{},
			},
		},
		{
			name:                 "004 Check absolute windows path with slashes",
			key:                  `C:/temp/t.torrent`,
			newTransferStructure: &libtorrent.TransferStructure{Opts: &options.Opts{}},
			expected: &libtorrent.TransferStructure{
				TorrentFilePath: `C:/temp/t.torrent`,
				TorrentFileName: "t.torrent",
				Opts:            &options.Opts{},
			},
		},
		{
			name:                 "005 Check absolute windows share path",
			key:                  `\\temp\t.torrent`,
			newTransferStructure: &libtorrent.TransferStructure{Opts: &options.Opts{}},
			expected: &libtorrent.TransferStructure{
				TorrentFilePath: `\\temp\t.torrent`,
				TorrentFileName: "t.torrent",
				Opts:            &options.Opts{},
			},
		},
		{
			name:                 "006 Check relative paths (torrent name)",
			key:                  "t.torrent",
			newTransferStructure: &libtorrent.TransferStructure{Opts: &options.Opts{BitDir: `C:\\temp`}},
			expected: &libtorrent.TransferStructure{
				TorrentFilePath: `C:\temp\t.torrent`,
				TorrentFileName: "t.torrent",
				Opts:            &options.Opts{BitDir: `C:\\temp`},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			HandleTorrentFilePath(testCase.newTransferStructure, testCase.key)
			equal := reflect.DeepEqual(testCase.expected, testCase.newTransferStructure)
			if !equal && !testCase.mustFail {
				changes, err := diff.Diff(testCase.newTransferStructure, testCase.expected, diff.DiscardComplexOrigin())
				if err != nil {
					t.Error(err.Error())
				}
				t.Fatalf("Unexpected error: structures aren't equal:\n Got: %#v\n Expect %#v\n Diff: %v\n", testCase.newTransferStructure, testCase.expected, spew.Sdump(changes))
			} else if equal && testCase.mustFail {
				t.Fatalf("Unexpected error: structures are equal, but they shouldn't\n Got: %#v\n", testCase.newTransferStructure)
			}
		})
	}
}

func TestPath(t *testing.T) {
	nts := libtorrent.CreateEmptyNewTransferStructure()
	err := helpers.DecodeTorrentFile("../../test/data/testfileset.torrent", nts)
	if err != nil {
		t.Fatalf("Can't decode torrent file with error: %v", err)
	}
}
