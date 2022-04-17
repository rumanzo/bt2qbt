package qBittorrentStructures

import (
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"testing"
)

func TestDecodeFastresumeFile(t *testing.T) {
	type PathJoinCase struct {
		name     string
		mustFail bool
		path     string
	}
	cases := []PathJoinCase{
		{
			name:     "001 not existing file",
			path:     "notexists.fastresume",
			mustFail: true,
		},
		{
			name: "002 testdir hybryd",
			path: "../../test/data/testdir_hybrid.fastresume",
		},
		{
			name: "003 testdir v1",
			path: "../../test/data/testdir_v1.fastresume",
		},
		{
			name: "004 testdir v2",
			path: "../../test/data/testdir_v2.fastresume",
		},
		{
			name: "005 single hybryd",
			path: "../../test/data/testfile1_single_hybrid.fastresume",
		},
		{
			name: "006 single v1",
			path: "../../test/data/testfile1_single_v1.fastresume",
		},
		{
			name: "007 single v2",
			path: "../../test/data/testfile1_single_v2.fastresume",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var decoded QBittorrentFastresume
			err := helpers.DecodeTorrentFile(testCase.path, &decoded)
			if err != nil && !testCase.mustFail {
				t.Fatalf("Unexpected error: %v", err)
			} else if err == nil && testCase.mustFail {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}
