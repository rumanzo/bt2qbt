package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rumanzo/bt2qbt/internal/options"
	"io/ioutil"
	"os"
)

func ProcessLabels(opts *options.Opts, newtags []string) error {
	categories := map[string]map[string]string{}

	// check if categories is new file. If it exists it must be unmarshaled. Default categories file contains only {}
	var categoriesIsNew bool
	file, err := os.OpenFile(opts.Categories, os.O_RDWR, 0644)
	if errors.Is(err, os.ErrNotExist) {
		categoriesIsNew = true
	} else if err != nil {
		return errors.New(fmt.Sprintf("Unexpected error while open categories.json. Error:\n%v\n", err))
	}

	if !categoriesIsNew {
		dataRaw, err := ioutil.ReadAll(file)
		if err != nil {
			return errors.New(fmt.Sprintf("Unexpected error while read categories.json. Error:\n%v\n", err))
		}

		err = file.Close()
		if err != nil {
			return errors.New(fmt.Sprintf("Can't close categories.json. Error:\n%v\n", err))
		}

		err = json.Unmarshal(dataRaw, &categories)
		if err != nil {
			return errors.New(fmt.Sprintf("Unexpected error while unmarshaling categories.json. Error:\n%v\n", err))
		}
	}

	for _, tag := range newtags {
		if _, ok := categories[tag]; !ok { // append only if key doesn't already exist
			categories[tag] = map[string]string{"save_path": ""}
		}
	}

	if !categoriesIsNew {
		err = os.Rename(opts.Categories, opts.Categories+".bak")
		if err != nil {
			return errors.New(fmt.Sprintf("Can't move categories.json to categories.bak. Error:\n%v\n", err))
		}
	}

	newCategories, err := json.Marshal(categories)
	if err != nil {
		return errors.New(fmt.Sprintf("Can't marshal categories. Error:\n%v\n", err))
	}

	err = ioutil.WriteFile(opts.Categories, newCategories, 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("Can't write categories.json. Error:\n%v\n", err))
	}

	return nil
}
