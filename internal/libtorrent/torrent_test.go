package libtorrent

import (
	"github.com/rumanzo/bt2qbt/pkg/qBittorrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"reflect"
	"testing"
)

func TestHandleTorrentFilePath(t *testing.T) {
	type SearchPathCase struct {
		name                string
		mustFail            bool
		newTorrentStructure *NewTorrentStructure
		key                 string
		expected            *NewTorrentStructure
	}

	cases := []SearchPathCase{
		{
			name: "001 Test torrent with windows single nofolder (original) path without replaces",
			newTorrentStructure: &NewTorrentStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: "D:\\torrents"},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
					},
				},
			},
			expected: &NewTorrentStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      "D:/torrents",
					SavePath:         "D:\\torrents",
					QBtContentLayout: "Original",
				},
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.newTorrentStructure.HandleSavePaths()
			equal := reflect.DeepEqual(testCase.expected.Fastresume, testCase.newTorrentStructure.Fastresume)
			if !equal && !testCase.mustFail {
				t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", testCase.newTorrentStructure.Fastresume, testCase.expected.Fastresume)
			} else if equal && testCase.mustFail {
				t.Fatal("Unexpected error: structures are equal, but they shouldn't\n", testCase.newTorrentStructure.Fastresume, testCase.expected.Fastresume)
			}
		})
	}
}
