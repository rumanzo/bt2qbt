package libtorrent

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/internal/replace"
	"github.com/rumanzo/bt2qbt/pkg/fileHelpers"
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

type TransferStructure struct {
	Fastresume      *qBittorrentStructures.QBittorrentFastresume `bencode:"-"`
	ResumeItem      *utorrentStructs.ResumeItem                  `bencode:"-"`
	TorrentFile     *torrentStructures.Torrent                   `bencode:"-"`
	TorrentFileRaw  map[string]interface{}                       `bencode:"-"`
	Opts            *options.Opts                                `bencode:"-"`
	TorrentFilePath string                                       `bencode:"-"`
	TorrentFileName string                                       `bencode:"-"`
	sizeAndPrio     [][]int64                                    `bencode:"-"`
	torrentFileList []string                                     `bencode:"-"`
	NumPieces       int64                                        `bencode:"-"`
	PieceLenght     int64                                        `bencode:"-"`
	Replace         []*replace.Replace                           `bencode:"-"`
	Targets         map[int64]string                             `bencode:"-"`
}

func CreateEmptyNewTransferStructure() TransferStructure {
	var transferStructure = TransferStructure{
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
		Opts:           &options.Opts{},
	}
	return transferStructure
}

func (transfer *TransferStructure) HandleCaption() {
	if transfer.ResumeItem.Caption != "" {
		transfer.Fastresume.QbtName = transfer.ResumeItem.Caption
	}
}

func (transfer *TransferStructure) HandleState() {
	if transfer.ResumeItem.Started == 0 {
		transfer.Fastresume.Paused = 1
		transfer.Fastresume.AutoManaged = 0
	} else {
		transfer.Fastresume.Paused = 0
		transfer.Fastresume.AutoManaged = 1
	}

}

func (transfer *TransferStructure) HandleTotalDownloaded() {
	if transfer.ResumeItem.CompletedOn == 0 {
		transfer.Fastresume.TotalDownloaded = 0
	} else {
		transfer.Fastresume.TotalDownloaded = transfer.ResumeItem.Downloaded
	}
}

func (transfer *TransferStructure) HandleCompleted() {
	if transfer.Fastresume.CompletedTime != 0 {
		transfer.Fastresume.LastSeenComplete = time.Now().Unix()
	} else {
		transfer.Fastresume.Unfinished = new([]interface{})
	}

}

func (transfer *TransferStructure) HandleTags() {
	if transfer.Opts.WithoutTags == false && transfer.ResumeItem.Labels != nil {
		for _, label := range transfer.ResumeItem.Labels {
			if label != "" {
				transfer.Fastresume.QbtTags = append(transfer.Fastresume.QbtTags, label)
			}
		}
	} else {
		transfer.Fastresume.QbtTags = []string{}
	}
}
func (transfer *TransferStructure) HandleLabels() {
	if transfer.Opts.WithoutLabels == false {
		transfer.Fastresume.QBtCategory = transfer.ResumeItem.Label
	} else {
		transfer.Fastresume.QBtCategory = ""
	}
}

func (transfer *TransferStructure) GetTrackers(trackers interface{}) {
	switch strct := trackers.(type) {
	case []interface{}:
		for _, st := range strct {
			transfer.GetTrackers(st)
		}
	case string:
		for _, str := range strings.Fields(strct) {
			transfer.Fastresume.Trackers = append(transfer.Fastresume.Trackers, []string{str})
		}

	}
}

func (transfer *TransferStructure) PrioConvert(src []byte) {
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
	transfer.Fastresume.FilePriority = newprio
}

func (transfer *TransferStructure) HandlePieces() {
	if transfer.Fastresume.Unfinished != nil {
		transfer.Fastresume.Pieces = transfer.FillWholePieces("0")
		if len(transfer.TorrentFile.Info.Files) > 0 {
			transfer.Fastresume.PiecePriority = transfer.FillPiecesParted()
		} else {
			transfer.Fastresume.PiecePriority = transfer.FillWholePieces("1")
		}
	} else {
		if len(transfer.TorrentFile.Info.Files) > 0 {
			transfer.Fastresume.Pieces = transfer.FillPiecesParted()
		} else {
			transfer.Fastresume.Pieces = transfer.FillWholePieces("1")
		}
		transfer.Fastresume.PiecePriority = transfer.Fastresume.Pieces
	}
}

func (transfer *TransferStructure) HandleSizes() {
	if len(transfer.TorrentFile.Info.Files) > 0 {
		var filelists [][]int64
		for num, file := range transfer.TorrentFile.Info.Files {
			var lenght, mtime int64
			var filestrings []string
			var mappedPath []string
			if file.PathUTF8 != nil {
				mappedPath = file.PathUTF8
			} else {
				mappedPath = file.Path
			}

			for n, f := range mappedPath {
				if len(mappedPath)-1 == n && len(transfer.Targets) > 0 {
					for index, rewrittenFileName := range transfer.Targets {
						if index == int64(num) {
							filestrings = append(filestrings, rewrittenFileName)
						}
					}
				} else {
					filestrings = append(filestrings, f)
				}
			}
			filename := strings.Join(filestrings, transfer.Opts.PathSeparator)
			transfer.torrentFileList = append(transfer.torrentFileList, filename)
			fullpath := transfer.ResumeItem.Path + transfer.Opts.PathSeparator + filename
			if n := transfer.Fastresume.FilePriority[num]; n != 0 {
				lenght = file.Length
				transfer.sizeAndPrio = append(transfer.sizeAndPrio, []int64{lenght, 1})
				mtime = helpers.Fmtime(fullpath)
			} else {
				lenght, mtime = 0, 0
				transfer.sizeAndPrio = append(transfer.sizeAndPrio,
					[]int64{file.Length, 0})
			}
			flenmtime := []int64{lenght, mtime}
			filelists = append(filelists, flenmtime)
		}
	}
}

func (transfer *TransferStructure) FillWholePieces(chr string) []byte {
	var newpieces = make([]byte, 0, transfer.NumPieces)
	nchr, _ := strconv.Atoi(chr)
	for i := int64(0); i < transfer.NumPieces; i++ {
		newpieces = append(newpieces, byte(nchr))
	}
	return newpieces
}

func (transfer *TransferStructure) GetHash() (hash string) {
	torinfo, _ := bencode.EncodeString(transfer.TorrentFileRaw["info"])
	h := sha1.New()
	io.WriteString(h, torinfo)
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

func (transfer *TransferStructure) FillPiecesParted() []byte {
	var newpieces = make([]byte, 0, transfer.NumPieces)
	var allocation [][]int64
	chrone, _ := strconv.Atoi("1")
	chrzero, _ := strconv.Atoi("0")
	offset := int64(0)
	for _, pair := range transfer.sizeAndPrio {
		allocation = append(allocation, []int64{offset + 1, offset + pair[0], pair[1]})
		offset = offset + pair[0]
	}
	for i := int64(0); i < transfer.NumPieces; i++ {
		belongs := false
		first, last := i*transfer.PieceLenght, (i+1)*transfer.PieceLenght
		for _, trio := range allocation {
			if (first >= trio[0]-transfer.PieceLenght && last <= trio[1]+transfer.PieceLenght) && trio[2] == 1 {
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

func (transfer *TransferStructure) HandleSavePaths() {
	// Original paths always ending with pathSeparator
	// SubFolder or NoSubfolder never have ending pathSeparator
	// qBtSavePath always has separator /, otherwise SavePath use os pathSeparator
	var torrentName string
	if transfer.TorrentFile.Info.NameUTF8 != "" {
		torrentName = transfer.TorrentFile.Info.NameUTF8
	} else {
		torrentName = transfer.TorrentFile.Info.Name
	}
	lastPathName := fileHelpers.Base(transfer.ResumeItem.Path)

	if len(transfer.TorrentFile.Info.Files) > 0 {
		if lastPathName == torrentName {
			transfer.Fastresume.QBtContentLayout = "Original"
			transfer.Fastresume.QbtSavePath = fileHelpers.CutLastPath(transfer.ResumeItem.Path, transfer.Opts.PathSeparator)
			if maxIndex := transfer.FindHighestIndexOfMappedFiles(); maxIndex >= 0 {
				transfer.Fastresume.MappedFiles = make([]string, maxIndex+1, maxIndex+1)
				for _, paths := range transfer.ResumeItem.Targets {
					index := paths[0].(int64)
					pathParts := make([]string, len(paths)-1, len(paths)-1)
					for num, part := range paths[1:] {
						pathParts[num] = part.(string)
					}
					// we have to append torrent name(from torrent file) at the top of path
					transfer.Fastresume.MappedFiles[index] = fileHelpers.Join(append([]string{torrentName}, pathParts...), transfer.Opts.PathSeparator)
				}
			}
			transfer.Fastresume.QbtSavePath = fileHelpers.CutLastPath(transfer.ResumeItem.Path, "/") + `/`
		} else {
			transfer.Fastresume.QBtContentLayout = "NoSubfolder"
			// NoSubfolder always has full mapped files
			// so we append all of them
			for _, filePath := range transfer.TorrentFile.Info.Files {
				var paths []string
				if len(filePath.PathUTF8) != 0 {
					paths = filePath.PathUTF8
				} else {
					paths = filePath.Path
				}
				transfer.Fastresume.MappedFiles = append(transfer.Fastresume.MappedFiles, fileHelpers.Join(paths, transfer.Opts.PathSeparator))
			}
			// and then doing remap if resumeItem contain target field
			if maxIndex := transfer.FindHighestIndexOfMappedFiles(); maxIndex >= 0 {
				for _, paths := range transfer.ResumeItem.Targets {
					index := paths[0].(int64)
					pathParts := make([]string, len(paths)-1, len(paths)-1)
					for num, part := range paths[1:] {
						pathParts[num] = part.(string)
					}
					transfer.Fastresume.MappedFiles[index] = fileHelpers.Join(pathParts, transfer.Opts.PathSeparator)
				}
			}
			transfer.Fastresume.QbtSavePath = fileHelpers.Normalize(transfer.ResumeItem.Path, "/")
		}
	} else {
		transfer.Fastresume.QBtContentLayout = "Original" // utorrent\bittorrent don't support create subfolders for torrents with single file
		if lastPathName == torrentName {
			transfer.Fastresume.QbtSavePath = fileHelpers.CutLastPath(transfer.ResumeItem.Path, `/`) + `/`
		} else {
			//it means that we have renamed path and targets item, and should have mapped files
			transfer.Fastresume.MappedFiles = []string{lastPathName}
			transfer.Fastresume.QbtSavePath = fileHelpers.CutLastPath(transfer.ResumeItem.Path, `/`) + `/`
		}
	}

	for _, pattern := range transfer.Replace {
		transfer.Fastresume.QbtSavePath = strings.ReplaceAll(transfer.Fastresume.QbtSavePath, pattern.From, pattern.To)
	}

	transfer.Fastresume.SavePath = fileHelpers.Normalize(transfer.Fastresume.QbtSavePath, transfer.Opts.PathSeparator)
	if transfer.Fastresume.QBtContentLayout == "Original" {
		transfer.Fastresume.SavePath += transfer.Opts.PathSeparator
	}
}

// just helper for creating mappedfiles
func (transfer *TransferStructure) FindHighestIndexOfMappedFiles() int64 {
	if resumeItem := transfer.ResumeItem; resumeItem.Targets != nil {
		lastElem := resumeItem.Targets[len(resumeItem.Targets)-1] // it must be like []interface{0, "path"}
		return lastElem[0].(int64)
	}
	return -1
}

func CreateReplaces(replaces []string) []*replace.Replace {
	r := []*replace.Replace{}
	for _, str := range replaces {
		patterns := strings.Split(str, ",")
		r = append(r, &replace.Replace{
			From: patterns[0],
			To:   patterns[1],
		})
	}
	return r
}
