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

func checknotexists(s string, tags []string) (bool, string) {
	for _, value := range tags {
		if value == s {
			return false, s
		}
	}
	return true, s
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

func encodetorrentfile(path string, newstructure *NewTorrentStructure) error {
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

type NewTorrentStructure struct {
	Active_time int64 `bencode:"active_time"`
	Added_time int64 `bencode:"added_time"`
	Announce_to_dht int64 `bencode:"announce_to_dht"`
	Announce_to_lsd int64 `bencode:"announce_to_lsd"`
	Announce_to_trackers int64 `bencode:"announce_to_trackers"`
	Auto_managed int64 `bencode:"auto_managed"`
	Banned_peers string `bencode:"banned_peers"`
	Banned_peers6 string `bencode:"banned_peers6"`
	Blockperpiece int64 `bencode:""blocks per piece"`
	Completed_time int64 `bencode:"completed_time"`
	Download_rate_limit int64 `bencode:"download_rate_limit"`
	Filesizes [][]int64 `bencode:"file sizes"`
	Fileformat string `bencode:"file-format"`
	Fileversion int64 `bencode:"file-version"`
	File_priority []int `bencode:"file_priority"`
	Finished_time int64 `bencode:"finished_time"`
	Infohash string `bencode:"info-hash"`
	Last_seen_complete int64 `bencode:"last_seen_complete"`
	Libtorrentversion string `bencode:"libtorrent-version"`
	Max_connections int64 `bencode:"max_connections"`
	Max_uploads int64 `bencode:"max_uploads"`
	Num_downloaded int64 `bencode:"num_downloaded"`
	Num_incomplete int64 `bencode:"num_incomplete"`
	Mapped_files []string `bencode:"mapped_files",omitempty"`
	Paused int64 `bencode:"paused"`
	Peers string `bencode:"peers"`
	Peers6 string `bencode:"peers6"`
	Pieces []byte `bencode:"pieces"`
	QbthasRootFolder int64 `bencode:"qBt-hasRootFolder"`
	Qbtcategory string `bencode:"qBt-category,omitempty"`
	Qbtname string `bencode:"qBt-name,omitempty"`
	QbtqueuePosition int `bencode:"qBt-queuePosition"`
	QbtratioLimit int64 `bencode:"qBt-ratioLimit"`
	QbtsavePath string `bencode:"qBt-savePath"`
	QbtseedStatus int64 `bencode:"qBt-seedStatus"`
	QbtseedingTimeLimit int64 `bencode:"qBt-seedingTimeLimit"`
	Qbttags []interface{} `bencode:"qBt-tags"`
	QbttempPathDisabled int64 `bencode:"qBt-tempPathDisabled"`
	Save_path string `bencode:"save_path"`
	Seed_mode int64 `bencode:"seed_mode"`
	Seeding_time int64 `bencode:"seeding_time"`
	Sequential_download int64 `bencode:"sequential_download"`
	Super_seeding int64 `bencode:"super_seeding"`
	Total_downloaded int64 `bencode:"total_downloaded"`
	Total_uploaded int64 `bencode:"total_uploaded"`
	Trackers [][]string `bencode:"trackers"`
	Upload_rate_limit int64 `bencode:"upload_rate_limit"`
	Unfinished *[]interface{} `bencode:"unfinished,omitempty"`
	}

func (newstructure *NewTorrentStructure) started(started int64) {
	if started == 0 {
		newstructure.Paused = 1
		newstructure.Auto_managed = 0
		newstructure.Announce_to_dht = 0
		newstructure.Announce_to_lsd = 0
		newstructure.Announce_to_trackers = 0
	} else {
		newstructure.Paused = 0
		newstructure.Auto_managed = 1
		newstructure.Announce_to_dht = 1
		newstructure.Announce_to_lsd = 1
		newstructure.Announce_to_trackers = 1
	}
}

func (newstructure *NewTorrentStructure) ifcompletedon() {
	if newstructure.Completed_time != 0 {
		newstructure.Last_seen_complete = int64(time.Now().Unix())
	} else {
		newstructure.Unfinished = new([]interface{})
	}
}
func (newstructure *NewTorrentStructure) iftags(with_tags *bool, labels []interface{}) {
	if *with_tags == true {
		newstructure.Qbttags  = labels
	} else {
		newstructure.Qbttags = new([]interface{})
	}
}
func (newstructure *NewTorrentStructure) iflabel(with_label *bool, label string) {
	if *with_label == true {
		newstructure.Qbtcategory  = label
	} else {
		newstructure.Qbtcategory = ""
	}
}
func (newstructure *NewTorrentStructure) gettrackers(trackers []interface{}) {
	for _, tracker := range trackers {
		newstructure.Trackers = append(newstructure.Trackers, []string{tracker.(string)})
	}
}

func (newstructure *NewTorrentStructure) prioconvert(src string) {
	var newprio []int
	for _, c := range []byte(src) {
		if i := int(c); (i == 0) || (i == 128) { // if not selected
			newprio = append(newprio, 0)
		} else if (i == 4) || (i == 8) { // if low or normal prio
			newprio = append(newprio, 1)
		} else if i == 12 { // if high prio
			newprio = append(newprio, 6)
		} else {
			newprio = append(newprio, 0)
		}
	}
	newstructure.File_priority = newprio
}

func (newstructure *NewTorrentStructure) fillpieces(sizeandprio *[][]int64, npieces *int64, piecelenght *int64, hasfiles *bool) {
	if newstructure.Unfinished != nil {
		newstructure.Pieces = newstructure.fillnothavefiles(npieces, "0")
	} else {
		if *hasfiles {
			newstructure.Pieces = newstructure.fillhavefiles(sizeandprio, npieces, piecelenght)
		} else {
			newstructure.Pieces = newstructure.fillnothavefiles(npieces, "1")
		}
	}
}

func (newstructure *NewTorrentStructure) fillsizes(torrentfile map[string]interface{}, torrentfilelist *[]string, path string, filesizes *float32, sizeandprio *[][]int64 ) {
	if files, ok := torrentfile["info"].(map[string]interface{})["files"]; ok {
		var filelists [][]int64
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
			*torrentfilelist = append(*torrentfilelist, filename)
			fullpath := path + string(os.PathSeparator) + filename
			*filesizes += float32(file.(map[string]interface{})["length"].(int64))
			if n := newstructure.File_priority[num]; n != 0 {
				lenght = file.(map[string]interface{})["length"].(int64)
				*sizeandprio = append(*sizeandprio, []int64{lenght, 1})
				mtime = fmtime(fullpath)
			} else {
				lenght, mtime = 0, 0
				*sizeandprio = append(*sizeandprio, []int64{file.(map[string]interface{})["length"].(int64), 0})
			}
			flenmtime := []int64{lenght, mtime}
			filelists = append(filelists, flenmtime)
		}
		newstructure.Filesizes = filelists
	} else {
		*filesizes = float32(torrentfile["info"].(map[string]interface{})["length"].(int64))
		newstructure.Filesizes = [][]int64{{torrentfile["info"].(map[string]interface{})["length"].(int64), fmtime(path)}}
	}
}

func (newstructure *NewTorrentStructure) fillnothavefiles(npieces *int64, chr string) []byte {
	var newpieces = make([]byte, 0, *npieces)
	for i := int64(0); i < *npieces; i++ {
		chr, _ := strconv.Atoi(chr)
		newpieces = append(newpieces, byte(chr))
	}
	return newpieces
}

func (newstructure *NewTorrentStructure) fillhavefiles(sizeandprio *[][]int64, npieces *int64, piecelenght *int64) []byte {
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

func (newstructure *NewTorrentStructure) fillsavepaths(torrentfile map[string]interface{}, path string, hasfiles *bool, torrentfilelist *[]string) {
	var torrentname string
	if name, ok := torrentfile["info"].(map[string]interface{})["name.utf-8"].(string); ok {
		torrentname = name
	} else {
		torrentname = torrentfile["info"].(map[string]interface{})["name"].(string)
	}
	origpath := path
	_, lastdirname := filepath.Split(strings.Replace(origpath, string(os.PathSeparator), "/", -1))
	if *hasfiles {
		if lastdirname == torrentname {
			newstructure.QbthasRootFolder = 1
			newstructure.Save_path = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.QbthasRootFolder = 0
			newstructure.Save_path = path + string(os.PathSeparator)
			newstructure.Mapped_files = *torrentfilelist
		}
	} else {
		if lastdirname == torrentname {
			newstructure.QbthasRootFolder = 0
			newstructure.Save_path = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.QbthasRootFolder = 0
			*torrentfilelist = append(*torrentfilelist, lastdirname)
			newstructure.Mapped_files = *torrentfilelist
			newstructure.Save_path = origpath[0 : len(origpath)-len(lastdirname)]
		}
	}
	newstructure.QbtsavePath = newstructure.Save_path
}

func logic(key string, value map[string]interface{}, bitdir *string, with_label *bool, with_tags *bool, qbitdir *string, comChannel chan string, position int) error {
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
	var npieces int64
	piecelenght := torrentfile["info"].(map[string]interface{})["piece length"].(int64)
	if ((filesizes / float32(piecelenght)) - float32((int64(filesizes) / piecelenght))) != 0 { // check fraction
		npieces = int64(filesizes)/torrentfile["info"].(map[string]interface{})["piece length"].(int64) + 1
	} else {
		npieces = int64(filesizes) / torrentfile["info"].(map[string]interface{})["piece length"].(int64)
	}

	newstructure := NewTorrentStructure{Active_time: 0, Added_time: 0, Announce_to_dht: 0, Announce_to_lsd: 0,
		Announce_to_trackers: 0, Auto_managed: 0, Blockperpiece: 0, Completed_time: 0, Download_rate_limit: -1,
		Fileformat: "libtorrent resume file", Fileversion: 1, Finished_time: 0,	Last_seen_complete: 0,
		Libtorrentversion: "1.1.6.0", Max_connections: 100, Max_uploads: 100, Num_downloaded: 0, Num_incomplete: 0,
		QbtqueuePosition: 1, QbtratioLimit: -2000, QbtseedStatus: 1, QbtseedingTimeLimit: -2, QbttempPathDisabled: 0,
		Seed_mode: 0, Seeding_time: 0, Sequential_download: 0, Super_seeding: 0, Total_downloaded: 0, Total_uploaded: 0,
		Upload_rate_limit: 0}
	newstructure.Active_time = value["runtime"].(int64)
	newstructure.Added_time = value["added_on"].(int64)
	newstructure.Completed_time = value["completed_on"].(int64)
	newstructure.Infohash = value["info"].(string)
	newstructure.Seeding_time = value["runtime"].(int64)
	newstructure.QbtqueuePosition = position
	newstructure.started(value["started"].(int64))
	newstructure.Finished_time = int64(time.Since(time.Unix(value["completed_on"].(int64), 0)).Minutes())
	newstructure.ifcompletedon()
	newstructure.Total_downloaded = value["downloaded"].(int64)
	newstructure.Total_uploaded = value["uploaded"].(int64)
	newstructure.Upload_rate_limit = value["upspeed"].(int64)
	newstructure.iftags(with_tags, value["labels"].([]string))
	newstructure.iflabel(with_tags, value["label"].(string))
	newstructure.gettrackers(value["trackers"].([]interface{}))
	newstructure.prioconvert(value["prio"].(string))
	newstructure.fillsizes(torrentfile, &torrentfilelist, value["path"].(string), &filesizes, &sizeandprio)
	newstructure.Blockperpiece = torrentfile["info"].(map[string]interface{})["piece length"].(int64) / value["blocksize"].(int64)
	newstructure.fillpieces(&sizeandprio, &npieces, &piecelenght, &hasfiles)
	newstructure.fillsavepaths(torrentfile, value["path"].(string), &hasfiles, &torrentfilelist)
	newbasename := gethash(torrentfile["info"].(map[string]interface{}))
	if err := encodetorrentfile(*qbitdir+newbasename+".fastresume", &newstructure); err != nil {
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
						if ok, tag := checknotexists(ASCIIconvert(label.(string)), newtags); ok {
							newtags = append(newtags, tag)
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
				if ok, t := checknotexists(tag, newtags); ok {
					newtags = append(newtags, t)
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
