package libtorrent

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/r3labs/diff/v2"
	_ "github.com/r3labs/diff/v2"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/qBittorrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"reflect"
	"testing"
)

// my fast test func
// todo remove this
//func TestRand(t *testing.T) {
//	test := map[string]interface{}{}
//	helpers.DecodeTorrentFile(`C:\Users\ruman\AppData\Roaming\uTorrent\resume.dat`, &test)
//	t.Fatal(spew.Sdump(test[`test1_renamed_qbt_мойфайл.txt.torrent`]))
//}

func TestTransferStructure_HandleSavePaths(t *testing.T) {
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
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test_torrent.txt`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: `test_torrent.txt`,
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `D:/torrents/`,
					SavePath:         `D:\torrents\`,
					QBtContentLayout: "Original",
				},
			},
		},
		{
			name: "002 Test torrent with windows single nofolder (original) path with replace",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test_torrent.txt`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent.txt",
					},
				},
				Opts: &options.Opts{PathSeparator: `\`, Replaces: []string{`D:/torrents,E:/newfolder`}},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `E:/newfolder/`,
					SavePath:         `E:\newfolder\`,
					QBtContentLayout: "Original",
				},
			},
		},
		{
			name: "003 Test torrent with windows single nofolder (original) path without replaces. NameUTF8",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test_torrent.txt`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						NameUTF8: "test_torrent.txt",
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `D:/torrents/`,
					SavePath:         `D:\torrents\`,
					QBtContentLayout: "Original",
				},
			},
		},
		{
			name: "004 Test torrent with windows single nofolder (original) path without replaces. Renamed File",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{
					Path: `D:\torrents\renamed_test_torrent.txt`,
					Targets: [][]interface{}{
						[]interface{}{
							0,
							"renamed_test_torrent.txt",
						},
					},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent.txt",
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `D:/torrents/`,
					SavePath:         `D:\torrents\`,
					QBtContentLayout: "Original",
					MappedFiles:      []string{`renamed_test_torrent.txt`},
				},
			},
		},
		{
			name: "005 Test torrent with windows single nofolder (original) path with replace to linux paths and linux sep. Renamed File",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{
					Path: `D:\torrents\renamed_test_torrent.txt`,
					Targets: [][]interface{}{
						[]interface{}{
							0,
							"renamed_test_torrent.txt",
						},
					},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent.txt",
					},
				},
				Opts: &options.Opts{PathSeparator: `/`, Replaces: []string{`D:/torrents,/mnt/d/torrents`}},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `/mnt/d/torrents/`,
					SavePath:         `/mnt/d/torrents/`,
					QBtContentLayout: "Original",
					MappedFiles:      []string{`renamed_test_torrent.txt`},
				},
			},
		},
		{
			name: "006 Test torrent with windows folder (original) path without replaces",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test_torrent`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `D:/torrents/`,
					SavePath:         `D:\torrents\`,
					QBtContentLayout: "Original",
				},
			},
		},
		{
			name: "007 Test torrent with windows folder (original) path with replace",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test_torrent`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`, Replaces: []string{`D:/torrents,E:/newfolder`}},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `E:/newfolder/`,
					SavePath:         `E:\newfolder\`,
					QBtContentLayout: "Original",
				},
			},
		},
		{
			name: "008 Test torrent with windows folder (original) path without replaces. NameUTF8",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test_torrent`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						NameUTF8: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `D:/torrents/`,
					SavePath:         `D:\torrents\`,
					QBtContentLayout: "Original",
				},
			},
		},
		{
			name: "009 Test torrent with windows folder (original) path without replaces. Renamed File",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{
					Path: `D:\torrents\test_torrent`,
					Targets: [][]interface{}{
						[]interface{}{
							int64(2),
							"renamed_test_torrent.txt",
						},
					},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `D:/torrents/`,
					SavePath:         `D:\torrents\`,
					QBtContentLayout: "Original",
					MappedFiles: []string{
						``,
						``,
						`test_torrent\renamed_test_torrent.txt`,
					},
				},
			},
		},
		{
			name: "010 Test torrent with windows folder (original) path with replace to linux paths and linux sep. Renamed File",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{
					Path: `D:\torrents\test_torrent`,
					Targets: [][]interface{}{
						[]interface{}{
							int64(2),
							"renamed_test_torrent.txt",
						},
					},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `/`, Replaces: []string{`D:/torrents,/mnt/d/torrents`}},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `/mnt/d/torrents/`,
					SavePath:         `/mnt/d/torrents/`,
					QBtContentLayout: "Original",
					MappedFiles: []string{
						``,
						``,
						`test_torrent/renamed_test_torrent.txt`,
					},
				},
			},
		},
		// NoSubfolder cases
		{
			name: "011 Test torrent with windows folder (NoSubfolder) path without replaces",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath: `D:/torrents/test`,
					SavePath:    `D:\torrents\test`,
					MappedFiles: []string{
						`dir1\file1.txt`,
						`dir2\file2.txt`,
						`file0.txt`,
					},
					QBtContentLayout: "NoSubfolder",
				},
			},
		},
		{
			name: "012 Test torrent with windows folder (NoSubfolder) path with replace",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`, Replaces: []string{`D:/torrents,E:/newfolder`}},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath: `E:/newfolder/test`,
					SavePath:    `E:\newfolder\test`,
					MappedFiles: []string{
						`dir1\file1.txt`,
						`dir2\file2.txt`,
						`file0.txt`,
					},
					QBtContentLayout: "NoSubfolder",
				},
			},
		},
		{
			name: "013 Test torrent with windows folder (NoSubfolder) path without replaces. NameUTF8",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						NameUTF8: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath: `D:/torrents/test`,
					SavePath:    `D:\torrents\test`,
					MappedFiles: []string{
						`dir1\file1.txt`,
						`dir2\file2.txt`,
						`file0.txt`,
					},
					QBtContentLayout: "NoSubfolder",
				},
			},
		},
		{
			name: "014 Test torrent with windows folder (NoSubfolder) path without replaces. Renamed File",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{
					Path: `D:\torrents\test`,
					Targets: [][]interface{}{
						[]interface{}{
							int64(2),
							"renamed_test_torrent.txt",
						},
					},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `D:/torrents/test`,
					SavePath:         `D:\torrents\test`,
					QBtContentLayout: "NoSubfolder",
					MappedFiles: []string{
						`dir1\file1.txt`,
						`dir2\file2.txt`,
						`renamed_test_torrent.txt`,
					},
				},
			},
		},
		{
			name: "015 Test torrent with windows folder (NoSubfolder) path with replace to linux paths and linux sep. Renamed File",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{
					Path: `D:\torrents\test`,
					Targets: [][]interface{}{
						[]interface{}{
							int64(2),
							"renamed_test_torrent.txt",
						},
					},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Path: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{Path: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `/`, Replaces: []string{`D:/torrents,/mnt/d/torrents`}},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `/mnt/d/torrents/test`,
					SavePath:         `/mnt/d/torrents/test`,
					QBtContentLayout: "NoSubfolder",
					MappedFiles: []string{
						`dir1/file1.txt`,
						`dir2/file2.txt`,
						`renamed_test_torrent.txt`,
					},
				},
			},
		},
		{
			name: "016 Test torrent with windows folder (NoSubfolder) path without replaces. TorrentPaths UTF8",
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: "test_torrent",
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{PathUTF8: []string{"dir1", "file1.txt"}},
							&torrentStructures.TorrentFile{PathUTF8: []string{"dir2", "file2.txt"}},
							&torrentStructures.TorrentFile{PathUTF8: []string{"file0.txt"}},
						},
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath: `D:/torrents/test`,
					SavePath:    `D:\torrents\test`,
					MappedFiles: []string{
						`dir1\file1.txt`,
						`dir2\file2.txt`,
						`file0.txt`,
					},
					QBtContentLayout: "NoSubfolder",
				},
			},
		},
		{
			name:     "017 Test torrent with windows single nofolder (original) path without replaces",
			mustFail: true,
			newTransferStructure: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{},
				ResumeItem: &utorrentStructs.ResumeItem{Path: `D:\torrents\test_torrent.txt`},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Name: `test_torrent.txt`,
					},
				},
				Opts: &options.Opts{PathSeparator: `\`},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					QbtSavePath:      `D:/torrents/`,
					SavePath:         `D:\torre`,
					QBtContentLayout: "Original",
				},
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.newTransferStructure.Opts != nil {
				replaces := CreateReplaces(testCase.newTransferStructure.Opts.Replaces)
				testCase.newTransferStructure.Replace = replaces
				testCase.expected.Replace = replaces
			}
			testCase.newTransferStructure.HandleSavePaths()
			equal := reflect.DeepEqual(testCase.expected.Fastresume, testCase.newTransferStructure.Fastresume)
			if !equal && !testCase.mustFail {
				changes, err := diff.Diff(testCase.newTransferStructure.Fastresume, testCase.expected.Fastresume, diff.DiscardComplexOrigin())
				if err != nil {
					t.Error(err.Error())
				}
				t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n Diff: %v\n", testCase.newTransferStructure.Fastresume, testCase.expected.Fastresume, spew.Sdump(changes))
			} else if equal && testCase.mustFail {
				t.Fatalf("Unexpected error: structures are equal, but they shouldn't\n Got: %v\n", spew.Sdump(testCase.newTransferStructure.Fastresume))
			}
		})
	}
}

func TestTransferStructure_HandlePieces(t *testing.T) {
	type FillPiecesPartedCase struct {
		name                 string
		mustFail             bool
		newTransferStructure *TransferStructure
		expected             *TransferStructure
	}

	cases := []FillPiecesPartedCase{
		{
			name: "001",
			newTransferStructure: &TransferStructure{
				NumPieces: 5,
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					FilePriority: []int64{1, 0, 1, 0, 0},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Length: 5},
							&torrentStructures.TorrentFile{Length: 5},
							&torrentStructures.TorrentFile{Length: 5},
							&torrentStructures.TorrentFile{Length: 5},
							&torrentStructures.TorrentFile{Length: 5},
						},
						PieceLength: 5,
					},
				},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					FilePriority: []int64{1, 0, 1, 0, 0},
					Pieces: []byte{
						byte(1),
						byte(0),
						byte(1),
						byte(0),
						byte(0),
					},
				},
			},
		},
		{
			name: "002",
			newTransferStructure: &TransferStructure{
				NumPieces: 5,
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					FilePriority: []int64{1, 0, 1},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Length: 13},
							&torrentStructures.TorrentFile{Length: 7},
							&torrentStructures.TorrentFile{Length: 5},
						},
						PieceLength: 5, // 25 total
					},
				},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					FilePriority: []int64{1, 0, 1},
					Pieces: []byte{
						byte(1),
						byte(1),
						byte(1),
						byte(0),
						byte(1),
					},
				},
			},
		},
		{
			name: "003",
			newTransferStructure: &TransferStructure{
				NumPieces: 5,
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					FilePriority: []int64{0, 1, 0},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Length: 9},
							&torrentStructures.TorrentFile{Length: 6},
							&torrentStructures.TorrentFile{Length: 10},
						},
						PieceLength: 5, // 25 total
					},
				},
			},
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					FilePriority: []int64{0, 1, 0},
					Pieces: []byte{
						byte(0),
						byte(1),
						byte(1),
						byte(0),
						byte(0),
					},
				},
			},
		},
		{
			name: "003 Mustfail",
			newTransferStructure: &TransferStructure{
				NumPieces: 5,
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					FilePriority: []int64{1, 0, 1},
				},
				TorrentFile: &torrentStructures.Torrent{
					Info: &torrentStructures.TorrentInfo{
						Files: []*torrentStructures.TorrentFile{
							&torrentStructures.TorrentFile{Length: 13},
							&torrentStructures.TorrentFile{Length: 7},
							&torrentStructures.TorrentFile{Length: 5},
						},
						PieceLength: 5, // 25 total
					},
				},
			},
			mustFail: true,
			expected: &TransferStructure{
				Fastresume: &qBittorrentStructures.QBittorrentFastresume{
					FilePriority: []int64{1, 0, 1},
					Pieces: []byte{
						byte(0),
						byte(1),
						byte(1),
						byte(0),
						byte(1),
					},
				},
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.newTransferStructure.HandlePieces()
			equal := reflect.DeepEqual(testCase.expected.Fastresume, testCase.newTransferStructure.Fastresume)
			if !equal && !testCase.mustFail {
				changes, err := diff.Diff(testCase.newTransferStructure.Fastresume, testCase.expected.Fastresume, diff.DiscardComplexOrigin())
				if err != nil {
					t.Error(err.Error())
				}
				t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n Diff: %v\n", testCase.newTransferStructure.Fastresume.Pieces, testCase.expected.Fastresume.Pieces, spew.Sdump(changes))
			} else if equal && testCase.mustFail {
				t.Fatalf("Unexpected error: structures are equal, but they shouldn't\n Got: %v\n", spew.Sdump(testCase.newTransferStructure.Fastresume))
			}
		})
	}
}
