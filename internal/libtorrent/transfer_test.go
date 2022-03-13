package libtorrent

import (
	"github.com/rumanzo/bt2qbt/pkg/qBittorrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"path/filepath"
	"reflect"
	"testing"
)

// my fast test func
// todo remove this
func TestRand(t *testing.T) {
	//path := `\\share\../test.file`
	//isAbs := filepath.IsAbs(path)
	//t.Fatalf("%v", isAbs)
	//var checkWindowsDiskPath = regexp.MustCompile(`^[A-Za-z]:\\\\`)
	t.Fatal(filepath.Join(``, `test`))
}

func TestHandleTorrentFilePath(t *testing.T) {
	type SearchPathCase struct {
		name                 string
		mustFail             bool
		newTransferStructure *TransferStructure
		key                  string
		expected             *TransferStructure
	}

	cases := []SearchPathCase{
		{
			name: "001 Test torrent with windows single nofolder (original) path without replaces",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: "D:\\torrents"},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
					},
				},
			},
			expected: &TransferStructure{
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
			testCase.newTransferStructure.HandleSavePaths()
			equal := reflect.DeepEqual(testCase.expected.Fastresume, testCase.newTransferStructure.Fastresume)
			if !equal && !testCase.mustFail {
				t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", testCase.newTransferStructure.Fastresume, testCase.expected.Fastresume)
			} else if equal && testCase.mustFail {
				t.Fatalf("Unexpected error: structures are equal, but they shouldn't\n Got: %#v\n", testCase.newTransferStructure.Fastresume)
			}
		})
	}
}
