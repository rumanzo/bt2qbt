package libtorrent

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/rumanzo/bt2qbt/internal/replace"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/qBittorrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"github.com/zeebo/bencode"
	"io"
	"strconv"
	"strings"
	"time"
)

type NewTorrentStructure struct {
	Fastresume      *qBittorrentStructures.QBittorrentFastresume `bencode:"-"`
	ResumeItem      *utorrentStructs.ResumeItem                  `bencode:"-"`
	TorrentFile     *torrentStructures.Torrent                   `bencode:"-"`
	TorrentFileRaw  map[string]interface{}                       `bencode:"-"`
	WithoutLabels   bool                                         `bencode:"-"`
	WithoutTags     bool                                         `bencode:"-"`
	TorrentFilePath string                                       `bencode:"-"`
	TorrentFileName string                                       `bencode:"-"`
	sizeAndPrio     [][]int64                                    `bencode:"-"`
	torrentFileList []string                                     `bencode:"-"`
	NumPieces       int64                                        `bencode:"-"`
	PieceLenght     int64                                        `bencode:"-"`
	Replace         []replace.Replace                            `bencode:"-"`
	Separator       string                                       `bencode:"-"`
	Targets         map[int64]string                             `bencode:"-"`
}

func CreateEmptyNewTorrentStructure() NewTorrentStructure {
	var newstructure = NewTorrentStructure{
		Fastresume: &qBittorrentStructures.QBittorrentFastresume{
			ActiveTime:          0,
			AddedTime:           0,
			Allocation:          "sparse",
			AutoManaged:         0,
			CompletedTime:       0,
			DownloadRateLimit:   -1,
			FileFormat:          "libtorrent resume file",
			FileVersion:         1,
			FinishedTime:        0,
			LastDownload:        0,
			LastSeenComplete:    0,
			LastUpload:          0,
			LibTorrentVersion:   "2.0.5.0",
			MaxConnections:      100,
			MaxUploads:          100,
			NumDownloaded:       0,
			NumIncomplete:       0,
			QbtRatioLimit:       -2000,
			QbtSeedStatus:       1,
			QbtSeedingTimeLimit: -2,
			SeedMode:            0,
			SeedingTime:         0,
			SequentialDownload:  0,
			SuperSeeding:        0,
			StopWhenReady:       0,
			TotalDownloaded:     0,
			TotalUploaded:       0,
			UploadRateLimit:     0,
			QbtName:             "",
		},
		TorrentFile:    &torrentStructures.Torrent{},
		TorrentFileRaw: map[string]interface{}{},
		ResumeItem:     &utorrentStructs.ResumeItem{},
		Targets:        map[int64]string{},
	}
	return newstructure
}

func (newStructure *NewTorrentStructure) HandleCaption() {
	if newStructure.ResumeItem.Caption != "" {
		newStructure.Fastresume.QbtName = newStructure.ResumeItem.Caption
	}
}

func (newStructure *NewTorrentStructure) HandleState() {
	if newStructure.ResumeItem.Started == 0 {
		newStructure.Fastresume.Paused = 1
		newStructure.Fastresume.AutoManaged = 0
	} else {
		newStructure.Fastresume.Paused = 0
		newStructure.Fastresume.AutoManaged = 1
	}

}

func (newStructure *NewTorrentStructure) HandleTotalDownloaded() {
	if newStructure.ResumeItem.CompletedOn == 0 {
		newStructure.Fastresume.TotalDownloaded = 0
	} else {
		newStructure.Fastresume.TotalDownloaded = newStructure.ResumeItem.Downloaded
	}
}

func (newStructure *NewTorrentStructure) HandleCompleted() {
	if newStructure.Fastresume.CompletedTime != 0 {
		newStructure.Fastresume.LastSeenComplete = time.Now().Unix()
	} else {
		newStructure.Fastresume.Unfinished = new([]interface{})
	}

}

func (newStructure *NewTorrentStructure) HandleTags() {
	if newStructure.WithoutTags == false && newStructure.ResumeItem.Labels != nil {
		for _, label := range newStructure.ResumeItem.Labels {
			if label != "" {
				newStructure.Fastresume.QbtTags = append(newStructure.Fastresume.QbtTags, label)
			}
		}
	} else {
		newStructure.Fastresume.QbtTags = []string{}
	}
}
func (newStructure *NewTorrentStructure) HandleLabels() {
	if newStructure.WithoutLabels == false {
		newStructure.Fastresume.QBtCategory = newStructure.ResumeItem.Label
	} else {
		newStructure.Fastresume.QBtCategory = ""
	}
}

func (newStructure *NewTorrentStructure) GetTrackers(trackers interface{}) {
	switch strct := trackers.(type) {
	case []interface{}:
		for _, st := range strct {
			newStructure.GetTrackers(st)
		}
	case string:
		for _, str := range strings.Fields(strct) {
			newStructure.Fastresume.Trackers = append(newStructure.Fastresume.Trackers, []string{str})
		}

	}
}

func (newStructure *NewTorrentStructure) PrioConvert(src []byte) {
	var newprio []int64
	for _, c := range src {
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
	newStructure.Fastresume.FilePriority = newprio
}

func (newStructure *NewTorrentStructure) HandlePieces() {
	if newStructure.Fastresume.Unfinished != nil {
		newStructure.Fastresume.Pieces = newStructure.FillWholePieces("0")
		if len(newStructure.TorrentFile.Info.Files) > 0 {
			newStructure.Fastresume.PiecePriority = newStructure.FillPiecesParted()
		} else {
			newStructure.Fastresume.PiecePriority = newStructure.FillWholePieces("1")
		}
	} else {
		if len(newStructure.TorrentFile.Info.Files) > 0 {
			newStructure.Fastresume.Pieces = newStructure.FillPiecesParted()
		} else {
			newStructure.Fastresume.Pieces = newStructure.FillWholePieces("1")
		}
		newStructure.Fastresume.PiecePriority = newStructure.Fastresume.Pieces
	}
}

func (newStructure *NewTorrentStructure) HandleSizes() {
	if len(newStructure.TorrentFile.Info.Files) > 0 {
		var filelists [][]int64
		for num, file := range newStructure.TorrentFile.Info.Files {
			var lenght, mtime int64
			var filestrings []string
			var mappedPath []string
			if file.PathUTF8 != nil {
				mappedPath = file.PathUTF8
			} else {
				mappedPath = file.Path
			}

			for n, f := range mappedPath {
				if len(mappedPath)-1 == n && len(newStructure.Targets) > 0 {
					for index, rewrittenFileName := range newStructure.Targets {
						if index == int64(num) {
							filestrings = append(filestrings, rewrittenFileName)
						}
					}
				} else {
					filestrings = append(filestrings, f)
				}
			}
			filename := strings.Join(filestrings, newStructure.Separator)
			newStructure.torrentFileList = append(newStructure.torrentFileList, filename)
			fullpath := newStructure.ResumeItem.Path + newStructure.Separator + filename
			if n := newStructure.Fastresume.FilePriority[num]; n != 0 {
				lenght = file.Length
				newStructure.sizeAndPrio = append(newStructure.sizeAndPrio, []int64{lenght, 1})
				mtime = helpers.Fmtime(fullpath)
			} else {
				lenght, mtime = 0, 0
				newStructure.sizeAndPrio = append(newStructure.sizeAndPrio,
					[]int64{file.Length, 0})
			}
			flenmtime := []int64{lenght, mtime}
			filelists = append(filelists, flenmtime)
		}
	}
}

func (newStructure *NewTorrentStructure) FillWholePieces(chr string) []byte {
	var newpieces = make([]byte, 0, newStructure.NumPieces)
	nchr, _ := strconv.Atoi(chr)
	for i := int64(0); i < newStructure.NumPieces; i++ {
		newpieces = append(newpieces, byte(nchr))
	}
	return newpieces
}

func (newStructure *NewTorrentStructure) GetHash() (hash string) {
	torinfo, _ := bencode.EncodeString(newStructure.TorrentFileRaw["info"])
	h := sha1.New()
	io.WriteString(h, torinfo)
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

func (newStructure *NewTorrentStructure) FillPiecesParted() []byte {
	var newpieces = make([]byte, 0, newStructure.NumPieces)
	var allocation [][]int64
	chrone, _ := strconv.Atoi("1")
	chrzero, _ := strconv.Atoi("0")
	offset := int64(0)
	for _, pair := range newStructure.sizeAndPrio {
		allocation = append(allocation, []int64{offset + 1, offset + pair[0], pair[1]})
		offset = offset + pair[0]
	}
	for i := int64(0); i < newStructure.NumPieces; i++ {
		belongs := false
		first, last := i*newStructure.PieceLenght, (i+1)*newStructure.PieceLenght
		for _, trio := range allocation {
			if (first >= trio[0]-newStructure.PieceLenght && last <= trio[1]+newStructure.PieceLenght) && trio[2] == 1 {
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

func (newStructure *NewTorrentStructure) HandleSavePaths() {
	var torrentname string
	if newStructure.TorrentFile.Info.NameUTF8 != "" {
		torrentname = newStructure.TorrentFile.Info.NameUTF8
	} else {
		torrentname = newStructure.TorrentFile.Info.Name
	}
	origpath := newStructure.ResumeItem.Path
	var dirpaths []string
	if contains := strings.Contains(origpath, "\\"); contains {
		dirpaths = strings.Split(origpath, "\\")
	} else {
		dirpaths = strings.Split(origpath, "/")
	}
	lastdirname := dirpaths[len(dirpaths)-1]
	if len(newStructure.TorrentFile.Info.Files) > 0 {
		if lastdirname == torrentname {
			newStructure.Fastresume.QBtContentLayout = "Original"
			newStructure.Fastresume.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
			if len(newStructure.Targets) > 0 {
				for _, path := range newStructure.torrentFileList {
					if len(path) > 0 {
						newStructure.Fastresume.MappedFiles = append(newStructure.Fastresume.MappedFiles, lastdirname+newStructure.Separator+path)
					} else {
						newStructure.Fastresume.MappedFiles = append(newStructure.Fastresume.MappedFiles, path)
					}
				}
			}
		} else {
			newStructure.Fastresume.QBtContentLayout = "NoSubfolder"
			newStructure.Fastresume.QbtSavePath = newStructure.ResumeItem.Path + newStructure.Separator
			newStructure.Fastresume.MappedFiles = newStructure.torrentFileList
		}
	} else {
		if lastdirname == torrentname {
			newStructure.Fastresume.QBtContentLayout = "Subfolder"
			newStructure.Fastresume.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newStructure.Fastresume.QBtContentLayout = "Original"
			newStructure.torrentFileList = append(newStructure.torrentFileList, lastdirname)
			newStructure.Fastresume.MappedFiles = newStructure.torrentFileList
			newStructure.Fastresume.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		}
	}
	for _, pattern := range newStructure.Replace {
		newStructure.Fastresume.QbtSavePath = strings.ReplaceAll(newStructure.Fastresume.QbtSavePath, pattern.From, pattern.To)
	}
	var oldsep string
	switch newStructure.Separator {
	case "\\":
		oldsep = "/"
	case "/":
		oldsep = "\\"
	}
	newStructure.Fastresume.QbtSavePath = strings.ReplaceAll(newStructure.Fastresume.QbtSavePath, oldsep, newStructure.Separator)
	newStructure.Fastresume.SavePath = strings.ReplaceAll(newStructure.Fastresume.QbtSavePath, "\\", "/")

	for num, entry := range newStructure.Fastresume.MappedFiles {
		newentry := strings.ReplaceAll(entry, oldsep, newStructure.Separator)
		if entry != newentry {
			newStructure.Fastresume.MappedFiles[num] = newentry
		}
	}
}
