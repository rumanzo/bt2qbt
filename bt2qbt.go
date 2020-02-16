package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-ini/ini"
	"github.com/juju/gnuflag"
	"github.com/zeebo/bencode"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Flags struct {
	bitDir, qBitDir, config    string
	withoutLabels, withoutTags bool
	replace                    string
}

type Replace struct {
	from, to string
}

type Channels struct {
	comChannel     chan string
	errChannel     chan string
	boundedChannel chan bool
}

type NewTorrentStructure struct {
	ActiveTime          int64          `bencode:"active_time"`
	AddedTime           int64          `bencode:"added_time"`
	AnnounceToDht       int64          `bencode:"announce_to_dht"`
	AnnounceToLsd       int64          `bencode:"announce_to_lsd"`
	AnnounceToTrackers  int64          `bencode:"announce_to_trackers"`
	AutoManaged         int64          `bencode:"auto_managed"`
	BannedPeers         string         `bencode:"banned_peers"`
	BannedPeers6        string         `bencode:"banned_peers6"`
	Blockperpiece       int64          `bencode:"blocks per piece"`
	CompletedTime       int64          `bencode:"completed_time"`
	DownloadRateLimit   int64          `bencode:"download_rate_limit"`
	Filesizes           [][]int64      `bencode:"file sizes"`
	Fileformat          string         `bencode:"file-format"`
	Fileversion         int64          `bencode:"file-version"`
	FilePriority        []int          `bencode:"file_priority"`
	FinishedTime        int64          `bencode:"finished_time"`
	Infohash            string         `bencode:"info-hash"`
	LastSeenComplete    int64          `bencode:"last_seen_complete"`
	Libtorrentversion   string         `bencode:"libtorrent-version"`
	MaxConnections      int64          `bencode:"max_connections"`
	MaxUploads          int64          `bencode:"max_uploads"`
	NumDownloaded       int64          `bencode:"num_downloaded"`
	NumIncomplete       int64          `bencode:"num_incomplete"`
	MappedFiles         []string       `bencode:"mapped_files,omitempty"`
	Paused              int64          `bencode:"paused"`
	Peers               string         `bencode:"peers"`
	Peers6              string         `bencode:"peers6"`
	Pieces              []byte         `bencode:"pieces"`
	QbthasRootFolder    int64          `bencode:"qBt-hasRootFolder"`
	Qbtcategory         string         `bencode:"qBt-category,omitempty"`
	Qbtname             string         `bencode:"qBt-name"`
	QbtqueuePosition    int            `bencode:"qBt-queuePosition"`
	QbtratioLimit       int64          `bencode:"qBt-ratioLimit"`
	QbtsavePath         string         `bencode:"qBt-savePath"`
	QbtseedStatus       int64          `bencode:"qBt-seedStatus"`
	QbtseedingTimeLimit int64          `bencode:"qBt-seedingTimeLimit"`
	Qbttags             []string       `bencode:"qBt-tags"`
	QbttempPathDisabled int64          `bencode:"qBt-tempPathDisabled"`
	SavePath            string         `bencode:"save_path"`
	SeedMode            int64          `bencode:"seed_mode"`
	SeedingTime         int64          `bencode:"seeding_time"`
	SequentialDownload  int64          `bencode:"sequential_download"`
	SuperSeeding        int64          `bencode:"super_seeding"`
	TotalDownloaded     int64          `bencode:"total_downloaded"`
	TotalUploaded       int64          `bencode:"total_uploaded"`
	Trackers            [][]string     `bencode:"trackers"`
	UploadRateLimit     int64          `bencode:"upload_rate_limit"`
	Unfinished          *[]interface{} `bencode:"unfinished,omitempty"`
	withoutLabels       bool
	withoutTags         bool
	hasFiles            bool
	torrentFilePath     string
	torrentfile         map[string]interface{}
	path                string
	fileSizes           int64
	sizeAndPrio         [][]int64
	torrentFileList     []string
	numPieces           int64
	pieceLenght         int64
	replace             []Replace
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

func ASCIIconvert(s string) string {
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
	if err := bencode.DecodeBytes(dat, &torrent); err != nil {
		return nil, err
	}
	return torrent, nil
}

func fmtime(path string) (mtime int64) {
	if fi, err := os.Stat(path); err != nil {
		return 0
	} else {
		mtime = fi.ModTime().Unix()
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

func (newstructure *NewTorrentStructure) started(started int64) {
	if started == 0 {
		newstructure.Paused = 1
		newstructure.AutoManaged = 0
		newstructure.AnnounceToDht = 0
		newstructure.AnnounceToLsd = 0
		newstructure.AnnounceToTrackers = 0
	} else {
		newstructure.Paused = 0
		newstructure.AutoManaged = 1
		newstructure.AnnounceToDht = 1
		newstructure.AnnounceToLsd = 1
		newstructure.AnnounceToTrackers = 1
	}
}

func (newstructure *NewTorrentStructure) ifcompletedon() {
	if newstructure.CompletedTime != 0 {
		newstructure.LastSeenComplete = time.Now().Unix()
	} else {
		newstructure.Unfinished = new([]interface{})
	}
}
func (newstructure *NewTorrentStructure) iftags(labels interface{}) {
	if newstructure.withoutTags == false && labels != nil {
		for _, label := range labels.([]interface{}) {
			if label != nil {
				newstructure.Qbttags = append(newstructure.Qbttags, label.(string))
			}
		}
	} else {
		newstructure.Qbttags = []string{}
	}
}
func (newstructure *NewTorrentStructure) iflabel(label interface{}) {
	if newstructure.withoutLabels == false {
		switch label.(type) {
		case nil:
			newstructure.Qbtcategory = ""
		case string:
			newstructure.Qbtcategory = label.(string)
		}
	} else {
		newstructure.Qbtcategory = ""
	}
}

func (newstructure *NewTorrentStructure) gettrackers(trackers interface{}) {
	switch strct := trackers.(type) {
	case []interface{}:
		for _, st := range strct {
			newstructure.gettrackers(st)
		}
	case string:
		for _, str := range strings.Fields(strct) {
			newstructure.Trackers = append(newstructure.Trackers, []string{str})
		}

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
	newstructure.FilePriority = newprio
}

func (newstructure *NewTorrentStructure) fillmissing() {
	newstructure.ifcompletedon()
	newstructure.fillsizes()
	newstructure.fillsavepaths()
	if newstructure.Unfinished != nil {
		newstructure.Pieces = newstructure.fillnothavefiles("0")
	} else {
		if newstructure.hasFiles {
			newstructure.Pieces = newstructure.fillhavefiles()
		} else {
			newstructure.Pieces = newstructure.fillnothavefiles("1")
		}
	}
}

func (newstructure *NewTorrentStructure) fillsizes() {
	newstructure.fileSizes = 0
	if newstructure.hasFiles {
		var filelists [][]int64
		for num, file := range newstructure.torrentfile["info"].(map[string]interface{})["files"].([]interface{}) {
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
			newstructure.torrentFileList = append(newstructure.torrentFileList, filename)
			fullpath := newstructure.path + string(os.PathSeparator) + filename
			newstructure.fileSizes += file.(map[string]interface{})["length"].(int64)
			if n := newstructure.FilePriority[num]; n != 0 {
				lenght = file.(map[string]interface{})["length"].(int64)
				newstructure.sizeAndPrio = append(newstructure.sizeAndPrio, []int64{lenght, 1})
				mtime = fmtime(fullpath)
			} else {
				lenght, mtime = 0, 0
				newstructure.sizeAndPrio = append(newstructure.sizeAndPrio,
					[]int64{file.(map[string]interface{})["length"].(int64), 0})
			}
			flenmtime := []int64{lenght, mtime}
			filelists = append(filelists, flenmtime)
		}
		newstructure.Filesizes = filelists
	} else {
		newstructure.fileSizes = newstructure.torrentfile["info"].(map[string]interface{})["length"].(int64)
		newstructure.Filesizes = [][]int64{{newstructure.torrentfile["info"].(map[string]interface{})["length"].(int64),
			fmtime(newstructure.path)}}
	}
}

func (newstructure *NewTorrentStructure) fillnothavefiles(chr string) []byte {
	var newpieces = make([]byte, 0, newstructure.numPieces)
	nchr, _ := strconv.Atoi(chr)
	for i := int64(0); i < newstructure.numPieces; i++ {
		newpieces = append(newpieces, byte(nchr))
	}
	return newpieces
}

func (newstructure *NewTorrentStructure) gethash() (hash string) {
	torinfo, _ := bencode.EncodeString(newstructure.torrentfile["info"].(map[string]interface{}))
	h := sha1.New()
	io.WriteString(h, torinfo)
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

func (newstructure *NewTorrentStructure) fillhavefiles() []byte {
	var newpieces = make([]byte, 0, newstructure.numPieces)
	var allocation [][]int64
	chrone, _ := strconv.Atoi("1")
	chrzero, _ := strconv.Atoi("0")
	offset := int64(0)
	for _, pair := range newstructure.sizeAndPrio {
		allocation = append(allocation, []int64{offset + 1, offset + pair[0], pair[1]})
		offset = offset + pair[0]
	}
	for i := int64(0); i < newstructure.numPieces; i++ {
		belongs := false
		first, last := i*newstructure.pieceLenght, (i+1)*newstructure.pieceLenght
		for _, trio := range allocation {
			if (first >= trio[0]-newstructure.pieceLenght && last <= trio[1]+newstructure.pieceLenght) && trio[2] == 1 {
				belongs = true
			}
		}
		if belongs {
			newpieces = append(newpieces, byte(chrone))
		} else {
			newpieces = append(newpieces, byte(chrzero))
		}
	}
	return newpieces
}

func (newstructure *NewTorrentStructure) fillsavepaths() {
	var torrentname string
	if name, ok := newstructure.torrentfile["info"].(map[string]interface{})["name.utf-8"].(string); ok {
		torrentname = name
	} else {
		torrentname = newstructure.torrentfile["info"].(map[string]interface{})["name"].(string)
	}
	origpath := newstructure.path
	_, lastdirname := filepath.Split(strings.Replace(origpath, string(os.PathSeparator), "/", -1))
	if newstructure.hasFiles {
		if lastdirname == torrentname {
			newstructure.QbthasRootFolder = 1
			newstructure.SavePath = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.QbthasRootFolder = 0
			newstructure.SavePath = newstructure.path + string(os.PathSeparator)
			newstructure.MappedFiles = newstructure.torrentFileList
		}
	} else {
		if lastdirname == torrentname {
			newstructure.QbthasRootFolder = 0
			newstructure.SavePath = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.QbthasRootFolder = 0
			newstructure.torrentFileList = append(newstructure.torrentFileList, lastdirname)
			newstructure.MappedFiles = newstructure.torrentFileList
			newstructure.SavePath = origpath[0 : len(origpath)-len(lastdirname)]
		}
	}
	if len(newstructure.replace) != 0 {
		for _, pattern := range newstructure.replace {
			newstructure.SavePath = strings.ReplaceAll(newstructure.SavePath, pattern.from, pattern.to)
		}
	}
	newstructure.QbtsavePath = newstructure.SavePath
}

func logic(key string, value map[string]interface{}, flags *Flags, chans *Channels, position int, wg *sync.WaitGroup) error {
	defer wg.Done()
	defer func() {
		<-chans.boundedChannel
	}()
	defer func() {
		if r := recover(); r != nil {
			chans.errChannel <- fmt.Sprintf(
				"Panic while processing torrent %v:\n======\nReason: %v.\nText panic:\n%v\n======",
				key, r, string(debug.Stack()))
		}
	}()
	var err error
	newstructure := NewTorrentStructure{ActiveTime: 0, AddedTime: 0, AnnounceToDht: 0, AnnounceToLsd: 0,
		AnnounceToTrackers: 0, AutoManaged: 0, CompletedTime: 0, DownloadRateLimit: -1,
		Fileformat: "libtorrent resume file", Fileversion: 1, FinishedTime: 0, LastSeenComplete: 0,
		Libtorrentversion: "1.1.6.0", MaxConnections: 100, MaxUploads: 100, NumDownloaded: 0, NumIncomplete: 0,
		QbtqueuePosition: 1, QbtratioLimit: -2000, QbtseedStatus: 1, QbtseedingTimeLimit: -2, QbttempPathDisabled: 0,
		SeedMode: 0, SeedingTime: 0, SequentialDownload: 0, SuperSeeding: 0, TotalDownloaded: 0, TotalUploaded: 0,
		UploadRateLimit: 0, Qbtname: "", withoutLabels: flags.withoutLabels, withoutTags: flags.withoutTags}
	if ok := filepath.IsAbs(key); ok {
		newstructure.torrentFilePath = key
	} else {
		newstructure.torrentFilePath = flags.bitDir + key
	}
	if _, err = os.Stat(newstructure.torrentFilePath); os.IsNotExist(err) {
		chans.errChannel <- fmt.Sprintf("Can't find torrent file %v for %v", newstructure.torrentFilePath, key)
		return err
	}
	newstructure.torrentfile, err = decodetorrentfile(newstructure.torrentFilePath)
	if err != nil {
		chans.errChannel <- fmt.Sprintf("Can't decode torrent file %v for %v", newstructure.torrentFilePath, key)
		return err
	}

	if flags.replace != "" {
		for _, str := range strings.Split(flags.replace, ";") {
			patterns := strings.Split(str, ",")
			newstructure.replace = append(newstructure.replace, Replace{
				from: patterns[0],
				to:   patterns[1],
			})
		}
	}

	if _, ok := newstructure.torrentfile["info"].(map[string]interface{})["files"]; ok {
		newstructure.hasFiles = true
	} else {
		newstructure.hasFiles = false
	}
	if value["path"].(string)[len(value["path"].(string))-1] == os.PathSeparator {
		newstructure.path = value["path"].(string)[:len(value["path"].(string))-1]
	} else {
		newstructure.path = value["path"].(string)
	}
	newstructure.ActiveTime = value["runtime"].(int64)
	newstructure.AddedTime = value["added_on"].(int64)
	newstructure.CompletedTime = value["completed_on"].(int64)
	newstructure.Infohash = value["info"].(string)
	newstructure.SeedingTime = value["runtime"].(int64)
	newstructure.QbtqueuePosition = position
	newstructure.started(value["started"].(int64))
	newstructure.FinishedTime = int64(time.Since(time.Unix(value["completed_on"].(int64), 0)).Minutes())
	newstructure.TotalDownloaded = value["downloaded"].(int64)
	newstructure.TotalUploaded = value["uploaded"].(int64)
	newstructure.UploadRateLimit = value["upspeed"].(int64)
	newstructure.iftags(value["labels"])
	if value["label"] != nil {
		newstructure.iflabel(value["label"].(string))
	} else {
		newstructure.iflabel("")
	}
	newstructure.gettrackers(value["trackers"])
	newstructure.prioconvert(value["prio"].(string))

	// https://libtorrent.org/manual-ref.html#fast-resume
	newstructure.Blockperpiece = newstructure.torrentfile["info"].(map[string]interface{})["piece length"].(int64) / 16 / 1024
	newstructure.pieceLenght = newstructure.torrentfile["info"].(map[string]interface{})["piece length"].(int64)

	/*
		pieces maps to a string whose length is a multiple of 20. It is to be subdivided into strings of length 20,
		each of which is the SHA1 hash of the piece at the corresponding index.
		http://www.bittorrent.org/beps/bep_0003.html
	*/
	newstructure.numPieces = int64(len(newstructure.torrentfile["info"].(map[string]interface{})["pieces"].(string))) / 20
	newstructure.fillmissing()
	newbasename := newstructure.gethash()

	if err = encodetorrentfile(flags.qBitDir+newbasename+".fastresume", &newstructure); err != nil {
		chans.errChannel <- fmt.Sprintf("Can't create qBittorrent fastresume file %v", flags.qBitDir+newbasename+".fastresume")
		return err
	}
	if err = copyfile(newstructure.torrentFilePath, flags.qBitDir+newbasename+".torrent"); err != nil {
		chans.errChannel <- fmt.Sprintf("Can't create qBittorrent torrent file %v", flags.qBitDir+newbasename+".torrent")
		return err
	}
	chans.comChannel <- fmt.Sprintf("Sucessfully imported %v", key)
	return nil
}

func main() {
	flags := Flags{}
	sep := string(os.PathSeparator)
	switch OS := runtime.GOOS; OS {
	case "windows":
		flags.bitDir = os.Getenv("APPDATA") + sep + "uTorrent" + sep
		flags.config = os.Getenv("APPDATA") + sep + "qBittorrent" + sep + "qBittorrent.ini"
		flags.qBitDir = os.Getenv("LOCALAPPDATA") + sep + "qBittorrent" + sep + "BT_backup" + sep
	case "linux":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		flags.bitDir = "/mnt/uTorrent/"
		flags.config = usr.HomeDir + sep + ".config" + sep + "qBittorrent" + sep + "qBittorrent.conf"
		flags.qBitDir = usr.HomeDir + sep + ".local" + sep + "share" + sep + "data" + sep + "qBittorrent" + sep + "BT_backup" + sep
	case "darwin":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		flags.bitDir = usr.HomeDir + sep + "Library" + sep + "Application Support" + sep + "uTorrent" + sep
		flags.config = usr.HomeDir + sep + ".config" + sep + "qBittorrent" + sep + "qbittorrent.ini"
		flags.qBitDir = usr.HomeDir + sep + "Library" + sep + "Application Support" + sep + "QBittorrent" + sep + "BT_backup" + sep
	}

	gnuflag.StringVar(&flags.bitDir, "source", flags.bitDir,
		"Source directory that contains resume.dat and torrents files")
	gnuflag.StringVar(&flags.bitDir, "s", flags.bitDir,
		"Source directory that contains resume.dat and torrents files")
	gnuflag.StringVar(&flags.qBitDir, "destination", flags.qBitDir,
		"Destination directory BT_backup (as default)")
	gnuflag.StringVar(&flags.qBitDir, "d", flags.qBitDir,
		"Destination directory BT_backup (as default)")
	gnuflag.StringVar(&flags.config, "qconfig", flags.config,
		"qBittorrent config files (for write tags)")
	gnuflag.StringVar(&flags.config, "c", flags.config,
		"qBittorrent config files (for write tags)")
	gnuflag.BoolVar(&flags.withoutLabels, "without-labels", false, "Do not export/import labels")
	gnuflag.BoolVar(&flags.withoutTags, "without-tags", false, "Do not export/import tags")
	gnuflag.StringVar(&flags.replace, "replace", "", "Replace paths.\n	"+
		"Delimiter for replaces - ;\n	"+
		"Delimiter for from/to - ,\n	Example: \"D:\\films,/home/user/films;\\,/\"\n	"+
		"If you use path separator different from you system, declare it mannually")
	gnuflag.Parse(true)

	if flags.replace != "" {
		for _, str := range strings.Split(flags.replace, ";") {
			patterns := strings.Split(str, ",")
			if len(patterns) < 2 {
				log.Println("Bad replace pattern")
				time.Sleep(30 * time.Second)
				os.Exit(1)
			}
		}
	}

	if flags.bitDir[len(flags.bitDir)-1] != os.PathSeparator {
		flags.bitDir += string(os.PathSeparator)
	}
	if flags.qBitDir[len(flags.qBitDir)-1] != os.PathSeparator {
		flags.qBitDir += string(os.PathSeparator)
	}

	if _, err := os.Stat(flags.bitDir); os.IsNotExist(err) {
		log.Println("Can't find uTorrent\\Bittorrent folder")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	if _, err := os.Stat(flags.qBitDir); os.IsNotExist(err) {
		log.Println("Can't find qBittorrent folder")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	resumefilepath := flags.bitDir + "resume.dat"
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
	if flags.withoutTags == false {
		if _, err := os.Stat(flags.config); os.IsNotExist(err) {
			fmt.Println("Can not read qBittorrent config file. Try run and close qBittorrent if you have not done" +
				" so already, or specify the path explicitly or do not import tags")
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
	}
	color.Green("It will be performed processing from directory %v to directory %v\n", flags.bitDir, flags.qBitDir)
	color.HiRed("Check that the qBittorrent is turned off and the directory %v and config %v is backed up.\n\n",
		flags.qBitDir, flags.config)
	fmt.Println("Press Enter to start")
	fmt.Scanln()
	log.Println("Started")
	totaljobs := len(resumefile)
	numjob := 1
	var oldtags string
	var newtags []string
	var wg sync.WaitGroup
	chans := Channels{comChannel: make(chan string, totaljobs),
		errChannel:     make(chan string, totaljobs),
		boundedChannel: make(chan bool, runtime.GOMAXPROCS(0)*2)}
	positionnum := 0
	for key, value := range resumefile {
		if key != ".fileguard" && key != "rec" {
			positionnum++
			if flags.withoutTags == false {
				if labels, ok := value.(map[string]interface{})["labels"]; ok {
					for _, label := range labels.([]interface{}) {
						if len(label.(string)) > 0 {
							if ok, tag := checknotexists(ASCIIconvert(label.(string)), newtags); ok {
								newtags = append(newtags, tag)
							}
						}
					}
				}
			}
			wg.Add(1)
			chans.boundedChannel <- true
			go logic(key, value.(map[string]interface{}), &flags, &chans, positionnum, &wg)
		} else {
			totaljobs--
		}
	}
	go func() {
		wg.Wait()
		close(chans.comChannel)
		close(chans.errChannel)
	}()
	for message := range chans.comChannel {
		fmt.Printf("%v/%v %v \n", numjob, totaljobs, message)
		numjob++
	}
	var waserrors bool
	for message := range chans.errChannel {
		fmt.Printf("%v/%v %v \n", numjob, totaljobs, message)
		waserrors = true
		numjob++
	}
	if flags.withoutTags == false {
		cfg, err := ini.Load(flags.config)
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
		cfg.SaveTo(flags.config)
	}
	fmt.Println()
	log.Println("Ended")
	if waserrors {
		log.Println("Not all torrents was processed")
	}
	fmt.Println("\nPress Enter to exit")
	fmt.Scanln()
}
