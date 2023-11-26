package helpers

import (
	"bufio"
	"bytes"
	"github.com/crazytyper/go-cesu8"
	"github.com/zeebo/bencode"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func ASCIIConvert(s string) string {
	var buffer bytes.Buffer
	for _, c := range s {
		if c > 127 {
			buffer.WriteString(`\x` + strconv.FormatUint(uint64(c), 16))
		} else {
			buffer.WriteString(string(c))
		}
	}
	return buffer.String()
}

// CheckExists return true and string if string exists in array, else false and string
func CheckExists(s string, arr []string) (bool, string) {
	for _, value := range arr {
		if value == s {
			return true, s
		}
	}
	return false, s
}

func DecodeTorrentFile(path string, decodeTo interface{}) error {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := bencode.DecodeBytes(dat, &decodeTo); err != nil {
		return err
	}
	return nil
}

func EncodeTorrentFile(path string, content interface{}) error {
	var err error
	var file *os.File
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err = os.Create(path)
		if err != nil {
			panic(err)
			return err
		}
		defer file.Close()
	} else {
		file, err = os.OpenFile(path, os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	bufferedWriter := bufio.NewWriter(file)

	enc := bencode.NewEncoder(bufferedWriter)
	if err = enc.Encode(content); err != nil {
		return err
	}
	err = bufferedWriter.Flush()
	if err := enc.Encode(content); err != nil {
		return err
	}
	return nil
}

func CopyFile(src string, dst string) error {
	originalFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer originalFile.Close()
	newFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer newFile.Close()
	if _, err := io.Copy(newFile, originalFile); err != nil {
		return err
	}
	if err := newFile.Sync(); err != nil {
		return err
	}
	return nil
}

func GetStrings(trackers interface{}) []string {
	ntrackers := []string{}
	switch strct := trackers.(type) {
	case []string:
		for _, str := range strct {
			ntrackers = append(ntrackers, strings.Fields(str)...)
		}
	case string:
		ntrackers = append(ntrackers, strings.Fields(strct)...)
	case []interface{}:
		for _, st := range strct {
			ntrackers = append(ntrackers, GetStrings(st)...)
		}
	}
	return ntrackers
}

func HandleCesu8(str string) string {
	if strings.Contains(str, "\xed\xa0") {
		return cesu8.DecodeString([]byte(str))
	}
	return str
}

// ReplaceAllSymbols Replace all symbols in set to replacer
func ReplaceAllSymbols(str string, set string, replacer string) string {
	for _, n := range set {
		str = strings.ReplaceAll(str, string(n), replacer)
	}
	return str
}
