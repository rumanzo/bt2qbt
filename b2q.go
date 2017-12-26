package main

import (
	"github.com/zeebo/bencode"
	"fmt"
	"io/ioutil"
	"os"
	//"github.com/davecgh/go-spew"
	//"reflect"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"strings"
	"sync"
)

type Torrent struct {
	name, path    string
	prio          []byte
	added         int64
	labels, files []string
}

func decodetorrentfile(path string) map[string]interface{} {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var torrent map[string]interface{}
	if err := bencode.DecodeBytes([]byte(dat), &torrent); err != nil {
		panic(err)
	}
	return torrent
}

func gethash(info interface{}) string {
	torinfo, _ := bencode.EncodeString(info.(map[string]interface{}))
	h := sha1.New()
	io.WriteString(h, torinfo)
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
}

func logic(key string, value interface{}, storrents *[]*Torrent, bitdir *string, wg *sync.WaitGroup) {
	defer wg.Done()
	local := value.(map[string]interface{})
	uncastedlabel := local["labels"].([]interface{})
	newlabels := make([]string, len(uncastedlabel), len(uncastedlabel)+1)
	for num, value := range uncastedlabel {
		if value == nil {
			value = "Empty"
		}
		newlabels[num] = value.(string)
	}
	torrentfile := decodetorrentfile(*bitdir + key)
	var files []string
	if val, ok := torrentfile["info"].(map[string]interface{})["files"].([]interface{}); ok {
		for _, i := range val {
			pathslice := i.(map[string]interface{})["path"]
			var newpath []string
			for _, path := range pathslice.([]interface{}) {
				newpath = append(newpath, path.(string))
			}
			files = append(files, strings.Join(newpath, "/"))
		}
	}
	ltorrent := Torrent{key, local["path"].(string), []byte(local["prio"].(string)),
		local["added_on"].(int64), newlabels, files}
	*storrents = append(*storrents, &ltorrent)
	fmt.Println(ltorrent.name, gethash(torrentfile["info"]))
}

func main() {
	var wg sync.WaitGroup
	bitdir := "C:/Users/rumanzo/AppData/Roaming/BitTorrent/"
	bitfile := bitdir + "resume.dat"

	torrent := decodetorrentfile(bitfile)

	var storrents []*Torrent
	for key, value := range torrent {
		if key != ".fileguard" && key != "rec" {
			wg.Add(1)
			go logic(key, value, &storrents, &bitdir, &wg)
		}
	}
	wg.Wait()
	fmt.Println(storrents[0])
}
