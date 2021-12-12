package libtorrent

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"github.com/rumanzo/bt2qbt/internal/replace"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/qBittorrentStructures"
	"github.com/zeebo/bencode"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type NewTorrentStructure struct {
	Fastresume      qBittorrentStructures.QBittorrentFastresume
	WithoutLabels   bool                   `bencode:"-"`
	WithoutTags     bool                   `bencode:"-"`
	HasFiles        bool                   `bencode:"-"`
	TorrentFilePath string                 `bencode:"-"`
	TorrentFile     map[string]interface{} `bencode:"-"`
	Path            string                 `bencode:"-"`
	fileSizes       int64                  `bencode:"-"`
	sizeAndPrio     [][]int64              `bencode:"-"`
	torrentFileList []string               `bencode:"-"`
	NumPieces       int64                  `bencode:"-"`
	PieceLenght     int64                  `bencode:"-"`
	Replace         []replace.Replace      `bencode:"-"`
	Separator       string                 `bencode:"-"`
	Targets         map[int64]string       `bencode:"-"`
}

func (newstructure *NewTorrentStructure) Started(started int64) {
	if started == 0 {
		newstructure.Fastresume.Paused = 1
		newstructure.Fastresume.AutoManaged = 0
	} else {
		newstructure.Fastresume.Paused = 0
		newstructure.Fastresume.AutoManaged = 1
	}
}

func (newstructure *NewTorrentStructure) IfCompletedOn() {
	if newstructure.Fastresume.CompletedTime != 0 {
		newstructure.Fastresume.LastSeenComplete = time.Now().Unix()
	} else {
		newstructure.Fastresume.Unfinished = new([]interface{})
	}
}
func (newstructure *NewTorrentStructure) IfTags(labels interface{}) {
	if newstructure.WithoutTags == false && labels != nil {
		for _, label := range labels.([]interface{}) {
			if label != nil {
				newstructure.Fastresume.QbtTags = append(newstructure.Fastresume.QbtTags, label.(string))
			}
		}
	} else {
		newstructure.Fastresume.QbtTags = []string{}
	}
}
func (newstructure *NewTorrentStructure) IfLabel(label interface{}) {
	if newstructure.WithoutLabels == false {
		switch label.(type) {
		case nil:
			newstructure.Fastresume.QBtCategory = ""
		case string:
			newstructure.Fastresume.QBtCategory = label.(string)
		}
	} else {
		newstructure.Fastresume.QBtCategory = ""
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
			newstructure.Fastresume.Trackers = append(newstructure.Fastresume.Trackers, []string{str})
		}

	}
}

func (newstructure *NewTorrentStructure) PrioConvert(src string) {
	var newprio []int64
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
	newstructure.Fastresume.FilePriority = newprio
}

func (newstructure *NewTorrentStructure) FillMissing() {
	newstructure.IfCompletedOn()
	newstructure.FillSizes()
	newstructure.FillSavePaths()
	if newstructure.Fastresume.Unfinished != nil {
		newstructure.Fastresume.Pieces = newstructure.FillWholePieces("0")
		if newstructure.HasFiles {
			newstructure.Fastresume.PiecePriority = newstructure.FillPiecesParted()
		} else {
			newstructure.Fastresume.PiecePriority = newstructure.FillWholePieces("1")
		}
	} else {
		if newstructure.HasFiles {
			newstructure.Fastresume.Pieces = newstructure.FillPiecesParted()
		} else {
			newstructure.Fastresume.Pieces = newstructure.FillWholePieces("1")
		}
		newstructure.Fastresume.PiecePriority = newstructure.Fastresume.Pieces
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
			if n := newstructure.Fastresume.FilePriority[num]; n != 0 {
				lenght = file.(map[string]interface{})["length"].(int64)
				newstructure.sizeAndPrio = append(newstructure.sizeAndPrio, []int64{lenght, 1})
				mtime = helpers.Fmtime(fullpath)
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
			newstructure.Fastresume.QBtContentLayout = "Original"
			newstructure.Fastresume.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
			if len(newstructure.Targets) > 0 {
				for _, path := range newstructure.torrentFileList {
					if len(path) > 0 {
						newstructure.Fastresume.MappedFiles = append(newstructure.Fastresume.MappedFiles, lastdirname+newstructure.Separator+path)
					} else {
						newstructure.Fastresume.MappedFiles = append(newstructure.Fastresume.MappedFiles, path)
					}
				}
			}
		} else {
			newstructure.Fastresume.QBtContentLayout = "NoSubfolder"
			newstructure.Fastresume.QbtSavePath = newstructure.Path + newstructure.Separator
			newstructure.Fastresume.MappedFiles = newstructure.torrentFileList
		}
	} else {
		if lastdirname == torrentname {
			newstructure.Fastresume.QBtContentLayout = "NoSubfolder"
			newstructure.Fastresume.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.Fastresume.QBtContentLayout = "NoSubfolder"
			newstructure.torrentFileList = append(newstructure.torrentFileList, lastdirname)
			newstructure.Fastresume.MappedFiles = newstructure.torrentFileList
			newstructure.Fastresume.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		}
	}
	for _, pattern := range newstructure.Replace {
		newstructure.Fastresume.QbtSavePath = strings.ReplaceAll(newstructure.Fastresume.QbtSavePath, pattern.From, pattern.To)
	}
	var oldsep string
	switch newstructure.Separator {
	case "\\":
		oldsep = "/"
	case "/":
		oldsep = "\\"
	}
	newstructure.Fastresume.QbtSavePath = strings.ReplaceAll(newstructure.Fastresume.QbtSavePath, oldsep, newstructure.Separator)
	newstructure.Fastresume.SavePath = strings.ReplaceAll(newstructure.Fastresume.QbtSavePath, "\\", "/")

	for num, entry := range newstructure.Fastresume.MappedFiles {
		newentry := strings.ReplaceAll(entry, oldsep, newstructure.Separator)
		if entry != newentry {
			newstructure.Fastresume.MappedFiles[num] = newentry
		}
	}
}

func EncodeTorrentFile(path string, newstructure *NewTorrentStructure) error {
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
