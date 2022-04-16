package torrentStructures

import (
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"testing"
)

func TestDecodeRealTorrents(t *testing.T) {
	type PathJoinCase struct {
		name     string
		mustFail bool
		path     string
	}
	cases := []PathJoinCase{
		{
			name:     "001 not existing file",
			path:     "notexists.torrent",
			mustFail: true,
		},
		{
			name: "002 existing file",
			path: "../../test/data/testfileset.torrent",
		},
		{
			name: "003 testdir hybryd",
			path: "../../test/data/testdir_hybrid.torrent",
		},
		{
			name: "004 testdir v1",
			path: "../../test/data/testdir_v1.torrent",
		},
		{
			name: "005 testdir v2",
			path: "../../test/data/testdir_v2.torrent",
		},
		{
			name: "006 single hybryd",
			path: "../../test/data/testfile1_single_hybrid.torrent",
		},
		{
			name: "007 single v1",
			path: "../../test/data/testfile1_single_v1.torrent",
		},
		{
			name: "008 single v2",
			path: "../../test/data/testfile1_single_v2.torrent",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var torrent Torrent
			err := helpers.DecodeTorrentFile(testCase.path, &torrent)
			if err != nil && !testCase.mustFail {
				t.Fatalf("Unexpected error: %v", err)
			} else if err == nil && testCase.mustFail {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}
