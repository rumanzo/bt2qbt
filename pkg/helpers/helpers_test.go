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
