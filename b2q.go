package main

import (
	"fmt"
	"github.com/zeebo/bencode"
	"io/ioutil"
	"os"
	//"reflect"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"strings"
	"sync"
	//"github.com/davecgh/go-spew/spew"
)

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

func logic(key string, value interface{}, bitdir *string, wg *sync.WaitGroup) {
	defer wg.Done()
	newstructure := map[string]interface{}{"active_time": new(int), "added_time": new(int), "announce_to_dht": new(int),
		"announce_to_lsd": new(int), "announce_to_trackers": new(int), "auto_managed": new(int),
		"banned_peers": new(string), "banned_peers6": new(string), "blocks per piece": new(int),
		"completed_time": new(int), "download_rate_limit": new(int), "file sizes": new([][]int),
		"file-format": new(int), "file-version": new(int), "file_priority": new([]int), "finished_time": new(int),
		"info-hash": new([]byte), "last_seen_complete": new(int), "libtorrent-version": new(string),
		"max_connections": new(int), "max_uploads": new(int), "num_complete": new(int), "num_downloaded": new(int),
		"num_incomplete": new(int), "paused": new(int), "peers": new(string), "peers6": new(string),
		"pieces": new([]byte), "qBt-category": new(string), "qBt-hasRootFolder": new(int), "qBt-name": new(string),
		"qBt-queuePosition": new(int), "qBt-ratioLimit": new(int), "qBt-savePath": new(string),
		"qBt-seedStatus": new(int), "qBt-seedingTimeLimit": new(int), "qBt-tags": new([]string),
		"qBt-tempPathDisabled": new(int), "save_path": new(string), "seed_mode": new(int), "seeding_time": new(int),
		"sequential_download": new(int), "super_seeding": new(int), "total_downloaded": new(int),
		"total_uploadedv": new(int), "trackers": new([][]string), "upload_rate_limit": new(int),
	}
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
	fmt.Println(newstructure)
}

func main() {
	var wg sync.WaitGroup
	bitdir := "C:/Users/rumanzo/AppData/Roaming/BitTorrent/"
	bitfile := bitdir + "resume.dat"

	torrent := decodetorrentfile(bitfile)

	for key, value := range torrent {
		if key != ".fileguard" && key != "rec" {
			wg.Add(1)
			go logic(key, value, &bitdir, &wg)
		}
	}
	wg.Wait()
}
