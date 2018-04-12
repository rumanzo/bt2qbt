package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-ini/ini"
	"github.com/zeebo/bencode"
	"io"
	"io/ioutil"
	"launchpad.net/gnuflag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ASCIIconvert(s string) (newstring string) {
	for _, c := range s {
		if c > 127 {
			newstring = fmt.Sprintf("%v\\x%x", newstring, c)
		} else {
			newstring = fmt.Sprintf("%v%v", newstring, string(c))
		}
	}
	return
}

func checknotexists(s string, tags []string) bool {
	for _, value := range tags {
		if value == s {
			return false
		}
	}
	return true
}

func decodetorrentfile(path string) (map[string]interface{}, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var torrent map[string]interface{}
	if err := bencode.DecodeBytes([]byte(dat), &torrent); err != nil {
		return nil, err
	}
	return torrent, nil
}

func encodetorrentfile(path string, newstructure map[string]interface{}) error {
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
	if err := enc.Encode(newstructure); err != nil {
		return err
	}
	bufferedWriter.Flush()
	return nil
}

func gethash(info map[string]interface{}) (hash string) {
	torinfo, _ := bencode.EncodeString(info)
	h := sha1.New()
	io.WriteString(h, torinfo)
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

func fillnothavefiles(npieces *int64, chr string) []byte {
	var newpieces = make([]byte, 0, *npieces)
	for i := int64(0); i < *npieces; i++ {
		chr, _ := strconv.Atoi(chr)
		newpieces = append(newpieces, byte(chr))
	}
	return newpieces
}

func fillhavefiles(sizeandprio *[][]int64, npieces *int64, piecelenght *int64) []byte {
	var newpieces = make([]byte, 0, *npieces)
	var allocation [][]int64
	offset := int64(0)
	for _, pair := range *sizeandprio {
		allocation = append(allocation, []int64{offset + 1, offset + pair[0], pair[1]})
		offset = offset + pair[0]
	}
	for i := int64(0); i < *npieces; i++ {
		belongs := false
		first, last := i**piecelenght, (i+1)**piecelenght
		for _, trio := range allocation {
			if (first >= trio[0]-*piecelenght && last <= trio[1]+*piecelenght) && trio[2] == 1 {
				belongs = true
			}
		}
		var chr int
		if belongs {
			chr, _ = strconv.Atoi("1")
		} else {
			chr, _ = strconv.Atoi("0")
		}
		newpieces = append(newpieces, byte(chr))
	}
	return newpieces
}

func prioconvert(src string) (newprio []int) {
	for _, c := range []byte(src) {
		if i := int(c); (i == 0) || (i == 128) { // if not selected
			newprio = append(newprio, 0)
		} else if (i == 4) || (i == 8) { // if low or normal prio
			newprio = append(newprio, 1)
		} else if i == 12 { // if high prio
			newprio = append(newprio, 6)
		}
	}
	return
}

func fmtime(path string) (mtime int64) {
	if fi, err := os.Stat(path); err != nil {
		return 0
	} else {
		mtime = int64(fi.ModTime().Unix())
		return
	}
}

func copyfile(src string, dst string) error {
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

func logic(key string, value map[string]interface{}, bitdir *string, with_label *bool, with_tags *bool, qbitdir *string, comChannel chan string, position int) error {
	newstructure := map[string]interface{}{"active_time": 0, "added_time": 0, "announce_to_dht": 0,
		"announce_to_lsd": 0, "announce_to_trackers": 0, "auto_managed": 0,
		"banned_peers": new(string), "banned_peers6": new(string), "blocks per piece": 0,
		"completed_time": 0, "download_rate_limit": -1, "file sizes": new([][]int64),
		"file-format": "libtorrent resume file", "file-version": 1, "file_priority": new([]int), "finished_time": 0,
		"info-hash": new([]byte), "last_seen_complete": 0, "libtorrent-version": "1.1.6.0",
		"max_connections": 100, "max_uploads": 100, "num_downloaded": 0,
		"num_incomplete": 0, "paused": new(int), "peers": new(string), "peers6": new(string),
		"pieces": new([]byte), "qBt-category": new(string), "qBt-name": new(string),
		"qBt-queuePosition": 1, "qBt-ratioLimit": -2000, "qBt-savePath": new(string),
		"qBt-seedStatus": 1, "qBt-seedingTimeLimit": -2, "qBt-tags": new([]string),
		"qBt-tempPathDisabled": 0, "save_path": new(string), "seed_mode": 0, "seeding_time": 0,
		"sequential_download": 0, "super_seeding": 0, "total_downloaded": 0,
		"total_uploaded": 0, "trackers": new([][]interface{}), "upload_rate_limit": 0,
	}
	torrentfilepath := *bitdir + key
	if _, err := os.Stat(torrentfilepath); os.IsNotExist(err) {
		comChannel <- fmt.Sprintf("Can't find torrent file %v for %v", torrentfilepath, key)
		return err
	}
	torrentfile, err := decodetorrentfile(torrentfilepath)
	if err != nil {
		comChannel <- fmt.Sprintf("Can't decode torrent file %v for %v", torrentfilepath, key)
		return err
	}
	newstructure["active_time"] = value["runtime"]
	newstructure["added_time"] = value["added_on"]
	newstructure["completed_time"] = value["completed_on"]
	newstructure["info-hash"] = value["info"]
	newstructure["qBt-tags"] = value["labels"]
	newstructure["seeding_time"] = value["runtime"]
	newstructure["qBt-queuePosition"] = position
	if value["started"].(int64) == int64(0) {
		newstructure["paused"] = 1
		newstructure["auto_managed"] = 0
		newstructure["announce_to_dht"] = 0
		newstructure["announce_to_lsd"] = 0
		newstructure["announce_to_trackers"] = 0
	} else {
		newstructure["paused"] = 0
		newstructure["auto_managed"] = 1
		newstructure["announce_to_dht"] = 1
		newstructure["announce_to_lsd"] = 1
		newstructure["announce_to_trackers"] = 1
	}
	newstructure["paused"] = 1
	newstructure["finished_time"] = int(time.Since(time.Unix(value["completed_on"].(int64), 0)).Minutes())
	if value["completed_on"] != 0 {
		newstructure["last_seen_complete"] = int(time.Now().Unix())
	} else {
		newstructure["unfinished"] = new([]interface{})
	}
	newstructure["total_downloaded"] = value["downloaded"]
	newstructure["total_uploaded"] = value["uploaded"]
	newstructure["upload_rate_limit"] = value["upspeed"]
	if *with_label == true {
		newstructure["qBt-category"] = value["label"]
	} else {
		newstructure["qBt-category"] = ""
	}
	if *with_tags == true {
		newstructure["qBt-tags"] = value["labels"]
	} else {
		newstructure["qBt-tags"] = ""
	}
	var trackers []interface{}
	for _, tracker := range value["trackers"].([]interface{}) {
		trackers = append(trackers, []interface{}{tracker})
	}
	newstructure["trackers"] = trackers
	newstructure["file_priority"] = prioconvert(value["prio"].(string))
	var hasfiles bool
	if _, ok := torrentfile["info"].(map[string]interface{})["files"]; ok {
		hasfiles = true
	} else {
		hasfiles = false
	}
	if value["path"].(string)[len(value["path"].(string))-1] == os.PathSeparator {
		value["path"] = value["path"].(string)[:len(value["path"].(string))-1]
	}
	filesizes := float32(0)
	var sizeandprio [][]int64
	var torrentfilelist []string
	if files, ok := torrentfile["info"].(map[string]interface{})["files"]; ok {
		var filelists []interface{}
		for num, file := range files.([]interface{}) {
			var lenght, mtime int64
			var filestrings []string
			if path, ok := file.(map[string]interface{})["path.utf-8"].([]interface{}); ok {
				for _, f := range path {
					filestrings = append(filestrings, f.(string))
				}
			} else {
				for _, f := range file.(map[string]interface{})["path"].([]interface{}) {
					filestrings = append(filestrings, f.(string))
				}
			}
			filename := strings.Join(filestrings, string(os.PathSeparator))
			torrentfilelist = append(torrentfilelist, filename)
			fullpath := value["path"].(string) + string(os.PathSeparator) + filename
			filesizes += float32(file.(map[string]interface{})["length"].(int64))
			if n := newstructure["file_priority"].([]int)[num]; n != 0 {
				lenght = file.(map[string]interface{})["length"].(int64)
				sizeandprio = append(sizeandprio, []int64{lenght, 1})
				mtime = fmtime(fullpath)
			} else {
				lenght, mtime = 0, 0
				sizeandprio = append(sizeandprio, []int64{file.(map[string]interface{})["length"].(int64), 0})
			}
			flenmtime := []int64{lenght, mtime}
			filelists = append(filelists, flenmtime)
		}
		newstructure["file sizes"] = filelists
	} else {
		filesizes = float32(torrentfile["info"].(map[string]interface{})["length"].(int64))
		newstructure["file sizes"] = [][]int64{{torrentfile["info"].(map[string]interface{})["length"].(int64), fmtime(value["path"].(string))}}
	}
	newstructure["blocks per piece"] = torrentfile["info"].(map[string]interface{})["piece length"].(int64) / value["blocksize"].(int64)
	var npieces int64
	piecelenght := torrentfile["info"].(map[string]interface{})["piece length"].(int64)
	if ((filesizes / float32(piecelenght)) - float32((int64(filesizes) / piecelenght))) != 0 { // check fraction
		npieces = int64(filesizes)/torrentfile["info"].(map[string]interface{})["piece length"].(int64) + 1
	} else {
		npieces = int64(filesizes) / torrentfile["info"].(map[string]interface{})["piece length"].(int64)
	}
	if _, ok := newstructure["unfinished"]; ok {
		newstructure["pieces"] = fillnothavefiles(&npieces, "0")
	} else {
		if hasfiles {
			newstructure["pieces"] = fillhavefiles(&sizeandprio, &npieces, &piecelenght)
		} else {
			newstructure["pieces"] = fillnothavefiles(&npieces, "1")
		}
	}
	var torrentname string
	if name, ok := torrentfile["info"].(map[string]interface{})["name.utf-8"].(string); ok {
		torrentname = name
	} else {
		torrentname = torrentfile["info"].(map[string]interface{})["name"].(string)
	}
	origpath := value["path"].(string)
	_, lastdirname := filepath.Split(strings.Replace(origpath, string(os.PathSeparator), "/", -1))
	if hasfiles {
		if lastdirname == torrentname {
			newstructure["qBt-hasRootFolder"] = 1
			newstructure["save_path"] = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure["qBt-hasRootFolder"] = 0
			newstructure["save_path"] = value["path"].(string) + string(os.PathSeparator)
			newstructure["mapped_files"] = torrentfilelist
		}
	} else {
		if lastdirname == torrentname {
			newstructure["qBt-hasRootFolder"] = 0
			newstructure["save_path"] = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure["qBt-hasRootFolder"] = 0
			torrentfilelist = append(torrentfilelist, lastdirname)
			newstructure["mapped_files"] = torrentfilelist
			newstructure["save_path"] = origpath[0 : len(origpath)-len(lastdirname)]
		}
	}
	newstructure["qBt-savePath"] = newstructure["save_path"]
	newbasename := gethash(torrentfile["info"].(map[string]interface{}))
	if err := encodetorrentfile(*qbitdir+newbasename+".fastresume", newstructure); err != nil {
		comChannel <- fmt.Sprintf("Can't create qBittorrent fastresume file %v", *qbitdir+newbasename+".fastresume")
		return err
	}
	if err := copyfile(*bitdir+key, *qbitdir+newbasename+".torrent"); err != nil {
		comChannel <- fmt.Sprintf("Can't create qBittorrent torrent file %v", *qbitdir+newbasename+".torrent")
		return err
	}
	comChannel <- fmt.Sprintf("Sucessfully imported %v", key)
	return nil
}

func main() {
	var bitdir, qbitdir, config string
	var with_label, with_tags bool = true, true
	var without_label, without_tags bool
	gnuflag.StringVar(&bitdir, "source", (os.Getenv("APPDATA") + "\\uTorrent\\"), "Source directory that contains resume.dat and torrents files")
	gnuflag.StringVar(&bitdir, "s", (os.Getenv("APPDATA") + "\\uTorrent\\"), "Source directory that contains resume.dat and torrents files")
	gnuflag.StringVar(&qbitdir, "destination", (os.Getenv("LOCALAPPDATA") + "\\qBittorrent\\BT_backup\\"), "Destination directory BT_backup (as default)")
	gnuflag.StringVar(&qbitdir, "d", (os.Getenv("LOCALAPPDATA") + "\\qBittorrent\\BT_backup\\"), "Destination directory BT_backup (as default)")
	gnuflag.StringVar(&config, "qconfig", (os.Getenv("APPDATA") + "\\qBittorrent\\qBittorrent.ini"), "qBittorrent config files (for write tags)")
	gnuflag.StringVar(&config, "c", (os.Getenv("APPDATA") + "\\qBittorrent\\qBittorrent.ini"), "qBittorrent config files (for write tags)")
	gnuflag.BoolVar(&without_label, "without-labels", false, "Do not export/import labels")
	gnuflag.BoolVar(&without_tags, "without-tags", false, "Do not export/import tags")
	gnuflag.Parse(true)

	if without_label {
		with_label = false
	}
	if without_tags {
		with_tags = false
	}

	if bitdir[len(bitdir)-1] != os.PathSeparator {
		bitdir += string(os.PathSeparator)
	}
	if qbitdir[len(qbitdir)-1] != os.PathSeparator {
		qbitdir += string(os.PathSeparator)
	}

	if _, err := os.Stat(bitdir); os.IsNotExist(err) {
		log.Println("Can't find uTorrent\\Bittorrent folder")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	if _, err := os.Stat(qbitdir); os.IsNotExist(err) {
		log.Println("Can't find qBittorrent folder")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	resumefilepath := bitdir + "resume.dat"
	if _, err := os.Stat(resumefilepath); os.IsNotExist(err) {
		log.Println("Can't find uTorrent\\Bittorrent resume file")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	resumefile, err := decodetorrentfile(resumefilepath)
	if err != nil {
		log.Println("Can't decode uTorrent\\Bittorrent resume file")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	if with_tags == true {
		if _, err := os.Stat(config); os.IsNotExist(err) {
			fmt.Println("Can not read qBittorrent config file. Try run and close qBittorrent if you have not done so already, or specify the path explicitly or do not import tags")
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
	}
	color.Green("It will be performed processing from directory %v to directory %v\n", bitdir, qbitdir)
	color.HiRed("Check that the qBittorrent is turned off and the directory %v and config %v is backed up.\n\n", qbitdir, config)
	fmt.Println("Press Enter to start")
	fmt.Scanln()
	fmt.Println("Started")
	totaljobs := len(resumefile) - 2
	numjob := 1
	var oldtags string
	var newtags []string
	comChannel := make(chan string, totaljobs)
	for key, value := range resumefile {
		if key != ".fileguard" && key != "rec" {
			if with_tags == true {
				for _, label := range value.(map[string]interface{})["labels"].([]interface{}) {
					if len(label.(string)) > 0 {
						if checknotexists(ASCIIconvert(label.(string)), newtags) {
							newtags = append(newtags, ASCIIconvert(label.(string)))
						}
					}
				}
			}
			go logic(key, value.(map[string]interface{}), &bitdir, &with_label, &with_tags, &qbitdir, comChannel, totaljobs)
		}
	}
	for message := range comChannel {
		fmt.Printf("%v/%v %v \n", numjob, totaljobs, message)
		numjob++
		if numjob-1 == totaljobs {
			break
		}
	}
	if with_tags == true {
		cfg, err := ini.Load(config)
		ini.PrettyFormat = false
		ini.PrettySection = false
		if err != nil {
			fmt.Println("Can not read qBittorrent config file. Try to specify the path explicitly or do not import tags")
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
		if _, err := cfg.GetSection("BitTorrent"); err != nil {
			cfg.NewSection("BitTorrent")

			//Dirty hack for section order. Sorry
			kv := cfg.Section("Network").KeysHash()
			cfg.DeleteSection("Network")
			cfg.NewSection("Network")
			for key, value := range kv {
				cfg.Section("Network").NewKey(key, value)
			}
			//End of dirty hack
		}
		if cfg.Section("BitTorrent").HasKey("Session\\Tags") {
			oldtags = cfg.Section("BitTorrent").Key("Session\\Tags").String()
			for _, tag := range strings.Split(oldtags, ", ") {
				if checknotexists(tag, newtags) {
					newtags = append(newtags, tag)
				}
			}
			cfg.Section("BitTorrent").Key("Session\\Tags").SetValue(strings.Join(newtags, ", "))
		} else {
			cfg.Section("BitTorrent").NewKey("Session\\Tags", strings.Join(newtags, ", "))
		}
		cfg.SaveTo(config)
	}
	fmt.Println("\nPress Enter to exit")
	fmt.Scanln()
}
