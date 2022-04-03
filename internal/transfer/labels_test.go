package transfer

import (
	"github.com/rumanzo/bt2qbt/internal/options"
	"io/ioutil"
	"os"
	"testing"
)

func TestProcessLabelsExisting(t *testing.T) {
	err := ioutil.WriteFile("../../test/categories_existing.json", []byte("{}"), 0755)
	if err != nil {
		t.Fatalf("Can't write empty categories test file. Err: %v", err.Error())
	}

	opts := &options.Opts{Categories: "../../test/categories_existing.json"}
	err = ProcessLabels(opts, []string{})
	if err != nil {
		t.Fatalf("Unexpecter error with handle categories. Err: %v", err.Error())
	}
	t.Cleanup(func() {
		err = os.Remove(opts.Categories + ".bak")
		if err != nil {
			t.Fatalf("It must exists bak file. Err: %v", err.Error())
		}
	})
}

func TestProcessLabelsNotExisting(t *testing.T) {
	opts := &options.Opts{Categories: "../../test/categories_not_existing.json"}
	os.Remove(opts.Categories)
	err := ProcessLabels(opts, []string{})
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Cleanup(func() {
		os.Remove(opts.Categories)
	})
}
