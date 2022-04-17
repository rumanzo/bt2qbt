package helpers

import (
	"reflect"
	"testing"
)

func TestGetStrings(t *testing.T) {
	testTrackers := []interface{}{
		"test1",
		"test2",
		[]interface{}{
			"test3",
			"test4",
			[]interface{}{
				"test5",
				"test6",
			},
		},
		[]interface{}{
			[]interface{}{"test7", "test8"},
		},
	}
	expect := []string{"test1", "test2", "test3", "test4", "test5", "test6", "test7", "test8"}
	trackers := GetStrings(testTrackers)
	if !reflect.DeepEqual(trackers, expect) {
		t.Fatalf("Unexpected error: opts isn't equal:\n Got: %#v\n Expect %#v\n", trackers, expect)
	}
}

func TestDecodeTorrentFile(t *testing.T) {
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
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var decoded interface{}
			err := DecodeTorrentFile(testCase.path, &decoded)
			if err != nil && !testCase.mustFail {
				t.Fatalf("Unexpected error: %v", err)
			} else if err == nil && testCase.mustFail {
				t.Fatalf("Test must fail, but it doesn't")
			}
		})
	}
}
