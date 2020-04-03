package libtorrent

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/rumanzo/bt2qbt/replace"
	"github.com/zeebo/bencode"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type NewTorrentStructure struct {
	ActiveTime          int64          `bencode:"active_time"`
	AddedTime           int64          `bencode:"added_time"`
	AnnounceToDht       int64          `bencode:"announce_to_dht"`
	AnnounceToLsd       int64          `bencode:"announce_to_lsd"`
	AnnounceToTrackers  int64          `bencode:"announce_to_trackers"`
	AutoManaged         int64          `bencode:"auto_managed"`
	BannedPeers         string         `bencode:"banned_peers"`
	BannedPeers6        string         `bencode:"banned_peers6"`
	BlockPerPiece       int64          `bencode:"blocks per piece"`
	CompletedTime       int64          `bencode:"completed_time"`
	DownloadRateLimit   int64          `bencode:"download_rate_limit"`
	FileSizes           [][]int64      `bencode:"file sizes"`
	FileFormat          string         `bencode:"file-format"`
	FileVersion         int64          `bencode:"file-version"`
	FilePriority        []int          `bencode:"file_priority"`
	FinishedTime        int64          `bencode:"finished_time"`
	InfoHash            string         `bencode:"info-hash"`
	LastSeenComplete    int64          `bencode:"last_seen_complete"`
	LibTorrentVersion   string         `bencode:"libtorrent-version"`
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
	QbtCategory         string         `bencode:"qBt-category,omitempty"`
	QbtName             string         `bencode:"qBt-name"`
	QbtQueuePosition    int            `bencode:"qBt-queuePosition"`
	QbtRatioLimit       int64          `bencode:"qBt-ratioLimit"`
	QbtSavePath         string         `bencode:"qBt-savePath"`
	QbtSeedStatus       int64          `bencode:"qBt-seedStatus"`
	QbtSeedingTimeLimit int64          `bencode:"qBt-seedingTimeLimit"`
	QbtTags             []string       `bencode:"qBt-tags"`
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
	WithoutLabels       bool
	WithoutTags         bool
	HasFiles            bool
	TorrentFilePath     string
	TorrentFile         map[string]interface{}
	Path                string
	fileSizes           int64
	sizeAndPrio         [][]int64
	torrentFileList     []string
	NumPieces           int64
	PieceLenght         int64
	Replace             []replace.Replace
}

func (newstructure *NewTorrentStructure) Started(started int64) {
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

func (newstructure *NewTorrentStructure) IfCompletedOn() {
	if newstructure.CompletedTime != 0 {
		newstructure.LastSeenComplete = time.Now().Unix()
	} else {
		newstructure.Unfinished = new([]interface{})
	}
}
func (newstructure *NewTorrentStructure) IfTags(labels interface{}) {
	if newstructure.WithoutTags == false && labels != nil {
		for _, label := range labels.([]interface{}) {
			if label != nil {
				newstructure.QbtTags = append(newstructure.QbtTags, label.(string))
			}
		}
	} else {
		newstructure.QbtTags = []string{}
	}
}
func (newstructure *NewTorrentStructure) IfLabel(label interface{}) {
	if newstructure.WithoutLabels == false {
		switch label.(type) {
		case nil:
			newstructure.QbtCategory = ""
		case string:
			newstructure.QbtCategory = label.(string)
		}
	} else {
		newstructure.QbtCategory = ""
	}
}

func (newstructure *NewTorrentStructure) GetTrackers(trackers interface{}) {
	switch strct := trackers.(type) {
	case []interface{}:
		for _, st := range strct {
			newstructure.GetTrackers(st)
		}
	case string:
		for _, str := range strings.Fields(strct) {
			newstructure.Trackers = append(newstructure.Trackers, []string{str})
		}

	}
}

func (newstructure *NewTorrentStructure) PrioConvert(src string) {
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

func (newstructure *NewTorrentStructure) FillMissing() {
	newstructure.IfCompletedOn()
	newstructure.FillSizes()
	newstructure.FillSavePaths()
	if newstructure.Unfinished != nil {
		newstructure.Pieces = newstructure.FillNotHaveFiles("0")
	} else {
		if newstructure.HasFiles {
			newstructure.Pieces = newstructure.FillHaveFiles()
		} else {
			newstructure.Pieces = newstructure.FillNotHaveFiles("1")
		}
	}
}

func (newstructure *NewTorrentStructure) FillSizes() {
	newstructure.fileSizes = 0
	if newstructure.HasFiles {
		var filelists [][]int64
		for num, file := range newstructure.TorrentFile["info"].(map[string]interface{})["files"].([]interface{}) {
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
			fullpath := newstructure.Path + string(os.PathSeparator) + filename
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
		newstructure.FileSizes = filelists
	} else {
		newstructure.fileSizes = newstructure.TorrentFile["info"].(map[string]interface{})["length"].(int64)
		newstructure.FileSizes = [][]int64{{newstructure.TorrentFile["info"].(map[string]interface{})["length"].(int64),
			fmtime(newstructure.Path)}}
	}
}

func (newstructure *NewTorrentStructure) FillNotHaveFiles(chr string) []byte {
	var newpieces = make([]byte, 0, newstructure.NumPieces)
	nchr, _ := strconv.Atoi(chr)
	for i := int64(0); i < newstructure.NumPieces; i++ {
		newpieces = append(newpieces, byte(nchr))
	}
	return newpieces
}

func (newstructure *NewTorrentStructure) GetHash() (hash string) {
	torinfo, _ := bencode.EncodeString(newstructure.TorrentFile["info"].(map[string]interface{}))
	h := sha1.New()
	io.WriteString(h, torinfo)
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

func (newstructure *NewTorrentStructure) FillHaveFiles() []byte {
	var newpieces = make([]byte, 0, newstructure.NumPieces)
	var allocation [][]int64
	chrone, _ := strconv.Atoi("1")
	chrzero, _ := strconv.Atoi("0")
	offset := int64(0)
	for _, pair := range newstructure.sizeAndPrio {
		allocation = append(allocation, []int64{offset + 1, offset + pair[0], pair[1]})
		offset = offset + pair[0]
	}
	for i := int64(0); i < newstructure.NumPieces; i++ {
		belongs := false
		first, last := i*newstructure.PieceLenght, (i+1)*newstructure.PieceLenght
		for _, trio := range allocation {
			if (first >= trio[0]-newstructure.PieceLenght && last <= trio[1]+newstructure.PieceLenght) && trio[2] == 1 {
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

func (newstructure *NewTorrentStructure) FillSavePaths() {
	var torrentname string
	if name, ok := newstructure.TorrentFile["info"].(map[string]interface{})["name.utf-8"].(string); ok {
		torrentname = name
	} else {
		torrentname = newstructure.TorrentFile["info"].(map[string]interface{})["name"].(string)
	}
	origpath := newstructure.Path
	_, lastdirname := filepath.Split(strings.Replace(origpath, string(os.PathSeparator), "/", -1))
	if newstructure.HasFiles {
		if lastdirname == torrentname {
			newstructure.QbthasRootFolder = 1
			newstructure.SavePath = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.QbthasRootFolder = 0
			newstructure.SavePath = newstructure.Path + string(os.PathSeparator)
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
	if len(newstructure.Replace) != 0 {
		for _, pattern := range newstructure.Replace {
			newstructure.SavePath = strings.ReplaceAll(newstructure.SavePath, pattern.From, pattern.To)
		}
	}
	newstructure.QbtSavePath = newstructure.SavePath
}

func fmtime(path string) (mtime int64) {
	if fi, err := os.Stat(path); err != nil {
		return 0
	} else {
		mtime = fi.ModTime().Unix()
		return
	}
}
