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

func TestEmojiCesu8(t *testing.T) {
	cesu8 := "normal_text \xed\xa0\xbc\xed\xb6\x95 normal_text \xed\xa0\xbd\xed\xba\x9c.txt.torrent"
	utf8 := "normal_text \xf0\x9f\x86\x95 normal_text \xf0\x9f\x9a\x9c.txt.torrent"
	if utf8 != HandleCesu8(cesu8) {
		t.Fatalf("Cesu8 to utf-8 transformation fail")
	}
}
func TestReplaceAllSymbols(t *testing.T) {
	type Case struct {
		name     string
		str      string
		set      string
		replacer string
		expected string
	}
	cases := []Case{
		{
			name:     "001 one symbol",
			str:      `qwerty`,
			set:      `qry`,
			replacer: `_`,
			expected: `_we_t_`,
		},
		{
			name:     "002 several replacer symbol",
			str:      `qwerty`,
			set:      `qry`,
			replacer: `AAA`,
			expected: `AAAweAAAtAAA`,
		},
		{
			name:     "003 several replacer symbol that exists in str",
			str:      `qwerty`,
			set:      `qry`,
			replacer: `qwerty`,
			expected: `qwertyweqwertytqwerty`,
		},
		{
			name:     "004 several replacer symbol that exists in str with special symbols",
			str:      `[qwerty]`,
			set:      `[qry]`,
			replacer: `qwerty`,
			expected: `qwertyqwertyweqwertytqwertyqwerty`,
		},
		{
			name:     "005 several replacer symbol that exists in str with special symbols",
			str:      `[qwerty]`,
			set:      `[qry]`,
			replacer: `[qwerty]`,
			expected: `[qwerty][qwerty]we[qwerty]t[qwerty][qwerty]`,
		},
		{
			name:     "006 emoji replace",
			str:      `qwer🚎y`,
			set:      `🚎`,
			replacer: `_`,
			expected: `qwer_y`,
		},
		{
			name:     "006 two emoji replace",
			str:      `qwer🚎y👍`,
			set:      `🚎😊`,
			replacer: `_`,
			expected: `qwer_y👍`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			replaced := ReplaceAllSymbols(testCase.str, testCase.set, testCase.replacer)
			if replaced != testCase.expected {
				t.Fatalf("Unexpected error:\nstr: %v set: %v replacer: %v\nGot: %v\nExpect %v\n", testCase.str, testCase.set, testCase.replacer, replaced, testCase.expected)
			}
		})
	}
}
