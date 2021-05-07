package libtorrent

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/rumanzo/bt2qbt/replace"
	"github.com/zeebo/bencode"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type NewTorrentStructure struct {
	ActiveTime          int64                  `bencode:"active_time"`
	AddedTime           int64                  `bencode:"added_time"`
	Allocation          string                 `bencode:"allocation"`
	AutoManaged         int64                  `bencode:"auto_managed"`
	CompletedTime       int64                  `bencode:"completed_time"`
	DownloadRateLimit   int64                  `bencode:"download_rate_limit"`
	FileFormat          string                 `bencode:"file-format"`
	FileVersion         int64                  `bencode:"file-version"`
	FilePriority        []int                  `bencode:"file_priority"`
	FinishedTime        int64                  `bencode:"finished_time"`
	HttpSeeds           []string               `bencode:"httpseeds"`
	InfoHash            string                 `bencode:"info-hash"`
	LastDownload        int64                  `bencode:"last_download"`
	LastSeenComplete    int64                  `bencode:"last_seen_complete"`
	LastUpload          int64                  `bencode:"last_upload"`
	LibTorrentVersion   string                 `bencode:"libtorrent-version"`
	MaxConnections      int64                  `bencode:"max_connections"`
	MaxUploads          int64                  `bencode:"max_uploads"`
	NumDownloaded       int64                  `bencode:"num_downloaded"`
	NumIncomplete       int64                  `bencode:"num_incomplete"`
	MappedFiles         []string               `bencode:"mapped_files,omitempty"`
	Paused              int64                  `bencode:"paused"`
	PiecePriority       []byte                 `bencode:"piece_priority"`
	Pieces              []byte                 `bencode:"pieces"`
	QbtCategory         string                 `bencode:"qBt-category,omitempty"`
	QbthasRootFolder    int64                  `bencode:"qBt-hasRootFolder"`
	QbtName             string                 `bencode:"qBt-name"`
	QbtQueuePosition    int                    `bencode:"qBt-queuePosition"`
	QbtRatioLimit       int64                  `bencode:"qBt-ratioLimit"`
	QbtSavePath         string                 `bencode:"qBt-savePath"`
	QbtSeedStatus       int64                  `bencode:"qBt-seedStatus"`
	QbtSeedingTimeLimit int64                  `bencode:"qBt-seedingTimeLimit"`
	QbtTags             []string               `bencode:"qBt-tags"`
	QbttempPathDisabled int64                  `bencode:"qBt-tempPathDisabled"`
	SavePath            string                 `bencode:"save_path"`
	SeedMode            int64                  `bencode:"seed_mode"`
	SeedingTime         int64                  `bencode:"seeding_time"`
	SequentialDownload  int64                  `bencode:"sequential_download"`
	StopWhenReady       int64                  `bencode:"stop_when_ready"`
	SuperSeeding        int64                  `bencode:"super_seeding"`
	TotalDownloaded     int64                  `bencode:"total_downloaded"`
	TotalUploaded       int64                  `bencode:"total_uploaded"`
	Trackers            [][]string             `bencode:"trackers"`
	UploadRateLimit     int64                  `bencode:"upload_rate_limit"`
	UrlList             int64                  `bencode:"url-list"`
	Unfinished          *[]interface{}         `bencode:"unfinished,omitempty"`
	WithoutLabels       bool                   `bencode:"-"`
	WithoutTags         bool                   `bencode:"-"`
	HasFiles            bool                   `bencode:"-"`
	TorrentFilePath     string                 `bencode:"-"`
	TorrentFile         map[string]interface{} `bencode:"-"`
	Path                string                 `bencode:"-"`
	fileSizes           int64                  `bencode:"-"`
	sizeAndPrio         [][]int64              `bencode:"-"`
	torrentFileList     []string               `bencode:"-"`
	NumPieces           int64                  `bencode:"-"`
	PieceLenght         int64                  `bencode:"-"`
	Replace             []replace.Replace      `bencode:"-"`
	Separator           string                 `bencode:"-"`
	Targets             map[int64]string       `bencode:"-"`
}

func (newstructure *NewTorrentStructure) Started(started int64) {
	if started == 0 {
		newstructure.Paused = 1
		newstructure.AutoManaged = 0
	} else {
		newstructure.Paused = 0
		newstructure.AutoManaged = 1
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
		} else if (i >= 1) && (i <= 8) { // if low or normal prio
			newprio = append(newprio, 1)
		} else if (i > 8) && (i <= 15) { // if high prio
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
		newstructure.Pieces = newstructure.FillWholePieces("0")
		if newstructure.HasFiles {
			newstructure.PiecePriority = newstructure.FillPiecesParted()
		} else {
			newstructure.PiecePriority = newstructure.FillWholePieces("1")
		}
	} else {
		if newstructure.HasFiles {
			newstructure.Pieces = newstructure.FillPiecesParted()
		} else {
			newstructure.Pieces = newstructure.FillWholePieces("1")
		}
		newstructure.PiecePriority = newstructure.Pieces
	}
}

func (newstructure *NewTorrentStructure) FillSizes() {
	newstructure.fileSizes = 0
	if newstructure.HasFiles {
		var filelists [][]int64
		for num, file := range newstructure.TorrentFile["info"].(map[string]interface{})["files"].([]interface{}) {
			var lenght, mtime int64
			var filestrings []string
			var mappedPath []interface{}
			if paths, ok := file.(map[string]interface{})["path.utf-8"].([]interface{}); ok {
				mappedPath = paths
			} else {
				mappedPath = file.(map[string]interface{})["path"].([]interface{})
			}

			for n, f := range mappedPath {
				if len(mappedPath)-1 == n && len(newstructure.Targets) > 0 {
					for index, rewrittenFileName := range newstructure.Targets {
						if index == int64(num) {
							filestrings = append(filestrings, rewrittenFileName)
						}
					}
				} else {
					filestrings = append(filestrings, f.(string))
				}
			}
			filename := strings.Join(filestrings, newstructure.Separator)
			newstructure.torrentFileList = append(newstructure.torrentFileList, filename)
			fullpath := newstructure.Path + newstructure.Separator + filename
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
	}
}

func (newstructure *NewTorrentStructure) FillWholePieces(chr string) []byte {
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

func (newstructure *NewTorrentStructure) FillPiecesParted() []byte {
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
	var dirpaths []string
	if contains := strings.Contains(origpath, "\\"); contains {
		dirpaths = strings.Split(origpath, "\\")
	} else {
		dirpaths = strings.Split(origpath, "/")
	}
	lastdirname := dirpaths[len(dirpaths)-1]
	if newstructure.HasFiles {
		if lastdirname == torrentname {
			newstructure.QbthasRootFolder = 1
			newstructure.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
			if len(newstructure.Targets) > 0 {
				for _, path := range newstructure.torrentFileList {
					if len(path) > 0 {
						newstructure.MappedFiles = append(newstructure.MappedFiles, lastdirname+newstructure.Separator+path)
					} else {
						newstructure.MappedFiles = append(newstructure.MappedFiles, path)
					}
				}
			}
		} else {
			newstructure.QbthasRootFolder = 0
			newstructure.QbtSavePath = newstructure.Path + newstructure.Separator
			newstructure.MappedFiles = newstructure.torrentFileList
		}
	} else {
		if lastdirname == torrentname {
			newstructure.QbthasRootFolder = 0
			newstructure.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.QbthasRootFolder = 0
			newstructure.torrentFileList = append(newstructure.torrentFileList, lastdirname)
			newstructure.MappedFiles = newstructure.torrentFileList
			newstructure.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		}
	}
	for _, pattern := range newstructure.Replace {
		newstructure.QbtSavePath = strings.ReplaceAll(newstructure.QbtSavePath, pattern.From, pattern.To)
	}
	var oldsep string
	switch newstructure.Separator {
	case "\\":
		oldsep = "/"
	case "/":
		oldsep = "\\"
	}
	newstructure.QbtSavePath = strings.ReplaceAll(newstructure.QbtSavePath, oldsep, newstructure.Separator)
	newstructure.SavePath = strings.ReplaceAll(newstructure.QbtSavePath, "\\", "/")

	for num, entry := range newstructure.MappedFiles {
		newentry := strings.ReplaceAll(entry, oldsep, newstructure.Separator)
		if entry != newentry {
			newstructure.MappedFiles[num] = newentry
		}
	}
}

func fmtime(path string) (mtime int64) {
	if fi, err := os.Stat(path); err != nil {
		return 0
	} else {
		mtime = fi.ModTime().Unix()
		return
	}
}
