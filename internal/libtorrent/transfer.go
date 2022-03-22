package libtorrent

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/internal/replace"
	"github.com/rumanzo/bt2qbt/pkg/fileHelpers"
	"github.com/rumanzo/bt2qbt/pkg/qBittorrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"github.com/zeebo/bencode"
	"io"
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
	NumPieces       int64                                        `bencode:"-"`
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

// recurstive function for searching trackers in resume item trackers
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

func (transfer *TransferStructure) HandlePriority() {
	for _, c := range transfer.ResumeItem.Prio {
		if i := int(c); (i == 0) || (i == 128) { // if priority not selected
			transfer.Fastresume.FilePriority = append(transfer.Fastresume.FilePriority, 0)
		} else if (i >= 1) && (i <= 8) { // if low or normal priority
			transfer.Fastresume.FilePriority = append(transfer.Fastresume.FilePriority, 1)
		} else if (i > 8) && (i <= 15) { // if high priority
			transfer.Fastresume.FilePriority = append(transfer.Fastresume.FilePriority, 6)
		} else {
			transfer.Fastresume.FilePriority = append(transfer.Fastresume.FilePriority, 0)
		}
	}
}

func (transfer *TransferStructure) GetHash() (hash string) {
	torinfo, _ := bencode.EncodeString(transfer.TorrentFileRaw["info"])
	h := sha1.New()
	io.WriteString(h, torinfo)
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

func (transfer *TransferStructure) HandlePieces() {
	if transfer.Fastresume.Unfinished != nil {
		transfer.FillWholePieces(0)
	} else {
		if len(transfer.TorrentFile.Info.Files) > 0 {
			transfer.FillPiecesParted()
		} else {
			transfer.FillWholePieces(1)
		}
	}
}

func (transfer *TransferStructure) FillWholePieces(piecePrio int) {
	transfer.Fastresume.Pieces = make([]byte, 0, transfer.NumPieces)
	for i := int64(0); i < transfer.NumPieces; i++ {
		transfer.Fastresume.Pieces = append(transfer.Fastresume.Pieces, byte(piecePrio))
	}
}

func (transfer *TransferStructure) FillPiecesParted() {
	transfer.Fastresume.Pieces = make([]byte, 0, transfer.NumPieces)

	// we count file offsets
	type Offset struct {
		firstOffset int64
		lastOffset  int64
	}
	var fileOffsets []Offset
	bytesLength := int64(0)
	for _, bytesFileLength := range transfer.TorrentFile.Info.Files {
		fileFirstOffset := bytesLength + 1
		bytesLength += bytesFileLength.Length
		fileLastOffset := bytesLength
		fileOffsets = append(fileOffsets, Offset{firstOffset: fileFirstOffset, lastOffset: fileLastOffset})
	}

	for i := int64(0); i < transfer.NumPieces; i++ {
		activePiece := false

		// we find fileOffset of pieces using piece length
		// https://libtorrent.org/manual-ref.html#fast-resume
		pieceOffset := Offset{
			firstOffset: i*transfer.TorrentFile.Info.PieceLength + 1,
			lastOffset:  (i + 1) * transfer.TorrentFile.Info.PieceLength,
		}

		// then we find indexes of the files that belongs to this piece
		for fileIndex, fileOffset := range fileOffsets {
			if fileOffset.firstOffset <= pieceOffset.lastOffset && fileOffset.lastOffset >= pieceOffset.firstOffset {
				// and if one of them have priority more than zero, we will append piece as completed
				if transfer.Fastresume.FilePriority[fileIndex] > 0 {
					activePiece = true
					break
				}
			}
		}

		if activePiece {
			transfer.Fastresume.Pieces = append(transfer.Fastresume.Pieces, byte(1))
		} else {
			transfer.Fastresume.Pieces = append(transfer.Fastresume.Pieces, byte(0))
		}
	}
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
