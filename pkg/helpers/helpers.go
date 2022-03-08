package helpers

import (
	"bufio"
	"bytes"
	"github.com/zeebo/bencode"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

func Fmtime(path string) (mtime int64) {
	if fi, err := os.Stat(path); err != nil {
		return 0
	} else {
		mtime = fi.ModTime().Unix()
		return
	}
}

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

// return true and string if string exists in array, else false and string
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
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Create(path)
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	bufferedWriter := bufio.NewWriter(file)
	enc := bencode.NewEncoder(bufferedWriter)
	if err := enc.Encode(content); err != nil {
		return err
	}
	bufferedWriter.Flush()
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
