package torrentStructures

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/r3labs/diff/v2"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"reflect"
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
		{
			name: "008 multi emoji prohibited symbols",
			path: "../../test/data/tests_slahes_emoji_ut_created.torrent",
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

func TestTorrent_GetFileList(t *testing.T) {
	type PathJoinCase struct {
		name     string
		path     string
		expected []string
		mustFail bool
	}
	cases := []PathJoinCase{
		{
			name:     "001 testdir v2 mustfail",
			path:     "../../test/data/testdir_v2.torrent",
			mustFail: true,
			expected: []string{},
		},
		{
			name: "002 testdir v2",
			path: "../../test/data/testdir_v2.torrent",
			expected: []string{
				"dir1/testfile1.txt",
				"dir2/testfile1.txt",
				"dir2/testfile2.txt",
				"dir3/testfile1.txt",
				"dir3/testfile2.txt",
				"dir3/testfile3.txt",
				"testfile1.txt",
				"testfile2.txt",
				"testfile3.txt",
			},
		},
		{
			name: "003 testdir v1",
			path: "../../test/data/testdir_v1.torrent",
			expected: []string{
				"testfile1.txt",
				"testfile2.txt",
				"testfile3.txt",
				"dir1/testfile1.txt",
				"dir2/testfile1.txt",
				"dir2/testfile2.txt",
				"dir3/testfile1.txt",
				"dir3/testfile2.txt",
				"dir3/testfile3.txt",
			},
		},
		{
			name: "004 testdir hybrid",
			path: "../../test/data/testdir_hybrid.torrent",
			expected: []string{
				"dir1/testfile1.txt",
				"dir2/testfile1.txt",
				"dir2/testfile2.txt",
				"dir3/testfile1.txt",
				"dir3/testfile2.txt",
				"dir3/testfile3.txt",
				"testfile1.txt",
				"testfile2.txt",
				"testfile3.txt",
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var torrent Torrent
			err := helpers.DecodeTorrentFile(testCase.path, &torrent)
			if err != nil {
				t.Fatalf("Unexpected error with decoding torrent file: %v", err)
			}
			list, _ := torrent.GetFileList()
			equal := reflect.DeepEqual(list, testCase.expected)
			if !equal && !testCase.mustFail {
				changes, err := diff.Diff(list, testCase.expected, diff.DiscardComplexOrigin())
				if err != nil {
					t.Error(err.Error())
				}
				t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n Diff: %v\n", list, testCase.expected, spew.Sdump(changes))
			} else if equal && testCase.mustFail {
				t.Fatalf("Unexpected error: structures are equal, but they shouldn't\n Got: %v\n", spew.Sdump(list))
			}
		})
	}
}

func TestTorrent_GetFileListWB(t *testing.T) {
	type PathJoinCase struct {
		name     string
		path     string
		expected []FilepathLength
		mustFail bool
	}
	cases := []PathJoinCase{
		{
			name:     "001 testdir v2 mustfail",
			path:     "../../test/data/testdir_v2.torrent",
			mustFail: true,
			expected: []FilepathLength{},
		},
		{
			name: "001 testdir v2",
			path: "../../test/data/testdir_v2.torrent",
			expected: []FilepathLength{
				FilepathLength{Path: "dir1/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir2/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir2/testfile2.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile2.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile3.txt", Length: 33},
				FilepathLength{Path: "testfile1.txt", Length: 33},
				FilepathLength{Path: "testfile2.txt", Length: 33},
				FilepathLength{Path: "testfile3.txt", Length: 33},
			},
		},
		{
			name: "003 testdir v1",
			path: "../../test/data/testdir_v1.torrent",
			expected: []FilepathLength{
				FilepathLength{Path: "testfile1.txt", Length: 33},
				FilepathLength{Path: "testfile2.txt", Length: 33},
				FilepathLength{Path: "testfile3.txt", Length: 33},
				FilepathLength{Path: "dir1/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir2/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir2/testfile2.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile2.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile3.txt", Length: 33},
			},
		},
		{
			name: "004 testdir hybrid",
			path: "../../test/data/testdir_hybrid.torrent",
			expected: []FilepathLength{
				FilepathLength{Path: "dir1/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir2/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir2/testfile2.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile2.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile3.txt", Length: 33},
				FilepathLength{Path: "testfile1.txt", Length: 33},
				FilepathLength{Path: "testfile2.txt", Length: 33},
				FilepathLength{Path: "testfile3.txt", Length: 33},
			},
		},
		{
			name: "004 testdir hybrid",
			path: "../../test/data/testdir_hybrid.torrent",
			expected: []FilepathLength{
				FilepathLength{Path: "dir1/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir2/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir2/testfile2.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile1.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile2.txt", Length: 33},
				FilepathLength{Path: "dir3/testfile3.txt", Length: 33},
				FilepathLength{Path: "testfile1.txt", Length: 33},
				FilepathLength{Path: "testfile2.txt", Length: 33},
				FilepathLength{Path: "testfile3.txt", Length: 33},
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var torrent Torrent
			err := helpers.DecodeTorrentFile(testCase.path, &torrent)
			if err != nil {
				t.Fatalf("Unexpected error with decoding torrent file: %v", err)
			}
			list, _ := torrent.GetFileListWB()
			equal := reflect.DeepEqual(list, testCase.expected)
			if !equal && !testCase.mustFail {
				changes, err := diff.Diff(list, testCase.expected, diff.DiscardComplexOrigin())
				if err != nil {
					t.Error(err.Error())
				}
				t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n Diff: %v\n", list, testCase.expected, spew.Sdump(changes))
			} else if equal && testCase.mustFail {
				t.Fatalf("Unexpected error: structures are equal, but they shouldn't\n Got: %v\n", spew.Sdump(list))
			}
		})
	}
}

func TestTorrent_GetFileListV1(t *testing.T) {
	type TestCases struct {
		name               string
		torrent            *Torrent
		expected           []FilepathLength
		expectedNormalized bool
		mustFail           bool
	}
	cases := []TestCases{
		{
			name: "001 emoji spaces at dir ends prohibited symbols",
			torrent: &Torrent{Info: &TorrentInfo{
				Files: []*TorrentFile{
					&TorrentFile{Path: []string{"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"file_with/slash.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ", "file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ", "file_with/slash.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_space ", "file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_space ", "file_with/slash.txt"}, Length: 5},
				},
			}},
			expectedNormalized: true,
			expected: []FilepathLength{
				{Path: "file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "file_with_slash.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_slash.txt", Length: 5},
			},
			mustFail: false,
		},
		{
			name: "002 mustFail wrong filename",
			torrent: &Torrent{Info: &TorrentInfo{
				Files: []*TorrentFile{
					&TorrentFile{Path: []string{"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"file_with/slash.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ", "file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ", "file_with/slash.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_space ", "file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_space ", "file_with/slash.txt"}, Length: 5},
				},
			}},
			expectedNormalized: true,
			expected: []FilepathLength{
				{Path: "file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "file_with_slash.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_space_/file__with_slash.txt", Length: 5},
			},
			mustFail: true,
		},
		{
			name: "003 emoji spaces at dir ends prohibited symbols mustfail due returned false normalized",
			torrent: &Torrent{Info: &TorrentInfo{
				Files: []*TorrentFile{
					&TorrentFile{Path: []string{"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"file_with/slash.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ", "file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ", "file_with/slash.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_space ", "file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir_with_space ", "file_with/slash.txt"}, Length: 5},
				},
			}},
			expectedNormalized: false,
			expected: []FilepathLength{
				{Path: "file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "file_with_slash.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_slash.txt", Length: 5},
			},
			mustFail: true,
		},
		{
			name: "004 normal torrent",
			torrent: &Torrent{Info: &TorrentInfo{
				Files: []*TorrentFile{
					&TorrentFile{Path: []string{"file.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir", "file.txt"}, Length: 5},
				},
			}},
			expectedNormalized: false,
			expected: []FilepathLength{
				{Path: "file.txt", Length: 5},
				{Path: "testdir/file.txt", Length: 5},
			},
			mustFail: false,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			filePathLength, normalized := getFileListV1(testCase.torrent)
			equal := reflect.DeepEqual(filePathLength, testCase.expected)
			if !equal && !testCase.mustFail {
				changes, err := diff.Diff(filePathLength, testCase.expected, diff.DiscardComplexOrigin())
				if err != nil {
					t.Error(err.Error())
				}
				t.Fatalf("Unexpected error: structures isn't equal:\n Got: %#v\n Expect %#v\n Diff: %v\n", filePathLength, testCase.expected, spew.Sdump(changes))
			}
			if normalized != testCase.expectedNormalized && !testCase.mustFail {
				t.Fatalf("Normalization expected %v, got %v", testCase.expectedNormalized, normalized)
			}
		})
	}
}
func TestTorrent_GetFileListV2(t *testing.T) {
	type TestCases struct {
		name               string
		torrent            *Torrent
		expected           []FilepathLength
		expectedNormalized bool
		mustFail           bool
	}
	cases := []TestCases{
		{
			name: "001 emoji spaces at dir ends prohibited symbols",
			torrent: &Torrent{Info: &TorrentInfo{
				FileTree: map[string]interface{}{
					"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
						``: map[string]interface{}{
							`length`:      int64(5),
							`pieces_root`: []byte{},
						},
					},
					"file_with/slash.txt": map[string]interface{}{
						``: map[string]interface{}{
							`length`:      int64(5),
							`pieces_root`: []byte{},
						},
					},
					"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ": map[string]interface{}{
						"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
						"file_with/slash.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
					},
					"testdir_with_space ": map[string]interface{}{
						"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
						"file_with/slash.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
					},
				},
			}},
			expectedNormalized: true,
			expected: []FilepathLength{
				{Path: "file_with_slash.txt", Length: 5},
				{Path: "file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
			},
			mustFail: false,
		},
		{
			name: "002 mustFail wrong filename",
			torrent: &Torrent{Info: &TorrentInfo{
				FileTree: map[string]interface{}{
					"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
						``: map[string]interface{}{
							`length`:      int64(5),
							`pieces_root`: []byte{},
						},
					},
					"file_with/slash.txt": map[string]interface{}{
						``: map[string]interface{}{
							`length`:      int64(5),
							`pieces_root`: []byte{},
						},
					},
					"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ": map[string]interface{}{
						"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
						"file_with/slash.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
					},
					"testdir_with_space ": map[string]interface{}{
						"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
						"file_with/slash.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
					},
				},
			}},
			expectedNormalized: true,
			expected: []FilepathLength{
				{Path: "file_with_slash.txt", Length: 5},
				{Path: "file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_emoji_ \xf0\x9f\x86\x95.txt", Length: 5},
			},
			mustFail: true,
		},
		{
			name: "003 emoji spaces at dir ends prohibited symbols mustfail due returned false normalized",
			torrent: &Torrent{Info: &TorrentInfo{
				FileTree: map[string]interface{}{
					"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
						``: map[string]interface{}{
							`length`:      int64(5),
							`pieces_root`: []byte{},
						},
					},
					"file_with/slash.txt": map[string]interface{}{
						``: map[string]interface{}{
							`length`:      int64(5),
							`pieces_root`: []byte{},
						},
					},
					"testdir_with_emoji_and_space \xed\xa0\xbc\xed\xb6\x95 ": map[string]interface{}{
						"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
						"file_with/slash.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
					},
					"testdir_with_space ": map[string]interface{}{
						"file_with_emoji \xed\xa0\xbc\xed\xb6\x95.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
						"file_with/slash.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
					},
				},
			}},
			expectedNormalized: false,
			expected: []FilepathLength{
				{Path: "file_with_slash.txt", Length: 5},
				{Path: "file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_emoji_and_space \xf0\x9f\x86\x95_/file_with_emoji \xf0\x9f\x86\x95.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_slash.txt", Length: 5},
				{Path: "testdir_with_space_/file_with_emoji_ \xf0\x9f\x86\x95.txt", Length: 5},
			},
			mustFail: true,
		},
		{
			name: "004 normal torrent",
			torrent: &Torrent{Info: &TorrentInfo{
				Files: []*TorrentFile{
					&TorrentFile{Path: []string{"file.txt"}, Length: 5},
					&TorrentFile{Path: []string{"testdir", "file.txt"}, Length: 5},
				},
				FileTree: map[string]interface{}{
					"file.txt": map[string]interface{}{
						``: map[string]interface{}{
							`length`:      int64(5),
							`pieces_root`: []byte{},
						},
					},
					"testdir": map[string]interface{}{
						"file.txt": map[string]interface{}{
							``: map[string]interface{}{
								`length`:      int64(5),
								`pieces_root`: []byte{},
							},
						},
					},
				},
			}},
			expectedNormalized: false,
			expected: []FilepathLength{
				{Path: "file.txt", Length: 5},
				{Path: "testdir/file.txt", Length: 5},
			},
			mustFail: false,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			filePathLength, normalized := getFileListV2(testCase.torrent.Info.FileTree)
			equal := reflect.DeepEqual(filePathLength, testCase.expected)
			if !equal && !testCase.mustFail {
				changes, err := diff.Diff(filePathLength, testCase.expected, diff.DiscardComplexOrigin())
				if err != nil {
					t.Error(err.Error())
				}
				t.Fatalf("Unexpected error: structures isn't equal:\n Got: %#v\nExpected: %#v\nDiff: %v\n", filePathLength, testCase.expected, spew.Sdump(changes))
			}
			if normalized != testCase.expectedNormalized && !testCase.mustFail {
				t.Fatalf("Normalization expected %v, got %v", testCase.expectedNormalized, normalized)
			}
		})
	}
}
