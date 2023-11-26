package transfer

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
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

//goland:noinspection GoNameStartsWithPackageName
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
	Magnet          bool                                         `bencode:"-"`
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
		transfer.Fastresume.QbtName = helpers.HandleCesu8(transfer.ResumeItem.Caption)
	}
}

// HandleState transfer torrents state.
// if torrent has several files and it doesn't complete downloaded (priority), it will be stopped
func (transfer *TransferStructure) HandleState() {
	if transfer.ResumeItem.Started == 0 {
		transfer.Fastresume.Paused = 1
		transfer.Fastresume.AutoManaged = 0
	} else {
		if len(transfer.TorrentFile.GetFileList()) > 1 {
			var parted bool
			for _, prio := range transfer.Fastresume.FilePriority {
				if prio == 0 {
					parted = true
					break
				}
			}
			if parted {
				transfer.Fastresume.Paused = 1
				transfer.Fastresume.AutoManaged = 0
				return
			}
		}
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
				transfer.Fastresume.QbtTags = append(transfer.Fastresume.QbtTags, helpers.HandleCesu8(label))
			}
		}
	} else {
		transfer.Fastresume.QbtTags = []string{}
	}
}
func (transfer *TransferStructure) HandleLabels() {
	if transfer.Opts.WithoutLabels == false {
		transfer.Fastresume.QBtCategory = helpers.HandleCesu8(transfer.ResumeItem.Label)
	} else {
		transfer.Fastresume.QBtCategory = ""
	}
}

var localTracker = regexp.MustCompile(`(http|udp)://\S+\.local\S*`)

func (transfer *TransferStructure) HandleTrackers() {
	trackers := helpers.GetStrings(transfer.ResumeItem.Trackers)
	trackersMap := map[string][]string{}
	var index string
	for _, tracker := range trackers {
		if localTracker.MatchString(tracker) {
			index = "local"
		} else {
			index = "main"
		}
		if _, ok := trackersMap[index]; ok {
			trackersMap[index] = append(trackersMap[index], helpers.HandleCesu8(tracker))
		} else {
			trackersMap[index] = []string{helpers.HandleCesu8(tracker)}
		}
	}
	if val, ok := trackersMap["main"]; ok {
		transfer.Fastresume.Trackers = append(transfer.Fastresume.Trackers, val)
	}
	if val, ok := trackersMap["local"]; ok {
		transfer.Fastresume.Trackers = append(transfer.Fastresume.Trackers, val)
	}
}

func (transfer *TransferStructure) HandlePriority() {
	if transfer.TorrentFile.IsV2OrHybryd() { // so we need get only odd
		trimmedPrio := make([]byte, 0, len(transfer.ResumeItem.Prio)/2)
		for i := 0; i < len(transfer.ResumeItem.Prio); i += 2 {
			trimmedPrio = append(trimmedPrio, transfer.ResumeItem.Prio[i])
		}
		transfer.ResumeItem.Prio = trimmedPrio
	}
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
		if len(transfer.TorrentFile.GetFileList()) > 0 {
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
	for _, file := range transfer.TorrentFile.GetFileListWB() { // need to adapt for torrents v2 version
		fileFirstOffset := bytesLength + 1
		bytesLength += file.Length
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

// we can't use these symbols on Windows systems, but can use in *nix
var prohibitedSymbols = regexp.MustCompilePOSIX(`[/:*?"<>|]`)

func (transfer *TransferStructure) HandleSavePaths() {
	// Original paths always ending with pathSeparator
	// SubFolder or NoSubfolder never have ending pathSeparator
	// qBtSavePath always has separator /, otherwise SavePath use os pathSeparator

	if transfer.Magnet {
		transfer.Fastresume.QBtContentLayout = "Original"
		transfer.Fastresume.QbtSavePath = fileHelpers.Normalize(helpers.HandleCesu8(transfer.ResumeItem.Path), "/")
	} else {
		var torrentName string
		var torrentNameOriginal string
		if transfer.TorrentFile.Info.NameUTF8 != "" {
			torrentName = helpers.HandleCesu8(transfer.TorrentFile.Info.NameUTF8)
			torrentNameOriginal = transfer.TorrentFile.Info.NameUTF8
		} else {
			torrentName = helpers.HandleCesu8(transfer.TorrentFile.Info.Name)
			torrentNameOriginal = transfer.TorrentFile.Info.Name
		}

		// transform windows prohibited symbols like libtorrent or utorrent do
		if transfer.Opts.PathSeparator == `\` {
			torrentName = prohibitedSymbols.ReplaceAllString(torrentName, `_`)
			torrentNameOriginal = prohibitedSymbols.ReplaceAllString(torrentNameOriginal, `_`)
		}
		lastPathName := fileHelpers.Base(helpers.HandleCesu8(transfer.ResumeItem.Path))

		if len(transfer.TorrentFile.GetFileList()) > 0 {
			var cesu8FilesExists bool
			for _, filePath := range transfer.TorrentFile.GetFileList() {
				cesuEncodedFilepath := helpers.HandleCesu8(filePath)
				if utf8.RuneCountInString(filePath) != utf8.RuneCountInString(cesuEncodedFilepath) {
					cesu8FilesExists = true
					break
				}
			}
			if lastPathName == torrentName && !cesu8FilesExists {
				transfer.Fastresume.QBtContentLayout = "Original"
				transfer.Fastresume.QbtSavePath = fileHelpers.CutLastPath(helpers.HandleCesu8(transfer.ResumeItem.Path), transfer.Opts.PathSeparator)
				if maxIndex := transfer.FindHighestIndexOfMappedFiles(); maxIndex >= 0 {
					transfer.Fastresume.MappedFiles = make([]string, maxIndex+1, maxIndex+1)
					for _, paths := range transfer.ResumeItem.Targets {
						index := paths[0].(int64)
						var pathParts []string
						if fileHelpers.IsAbs(helpers.HandleCesu8(paths[1].(string))) {
							pathParts = []string{fileHelpers.Normalize(helpers.HandleCesu8(paths[1].(string)), transfer.Opts.PathSeparator)}
							// if path is absolute just normalize it
							transfer.Fastresume.MappedFiles[index] = fileHelpers.Join(pathParts, transfer.Opts.PathSeparator)
						} else {
							pathParts = make([]string, len(paths)-1, len(paths)-1)
							for num, part := range paths[1:] {
								pathParts[num] = helpers.HandleCesu8(part.(string))
							}
							// we have to append torrent name(from torrent file) at the top of path
							transfer.Fastresume.MappedFiles[index] = fileHelpers.Join(append([]string{torrentName}, pathParts...), transfer.Opts.PathSeparator)
						}
					}
				}
				transfer.Fastresume.QbtSavePath = fileHelpers.CutLastPath(helpers.HandleCesu8(transfer.ResumeItem.Path), "/")
				if string(transfer.Fastresume.QbtSavePath[len(transfer.Fastresume.QbtSavePath)-1]) != `/` {
					transfer.Fastresume.QbtSavePath += `/`
				}
			} else {
				transfer.Fastresume.QBtContentLayout = "NoSubfolder"
				// NoSubfolder always has full mapped files
				// so we append all of them
				for _, filePath := range transfer.TorrentFile.GetFileList() {
					transfer.Fastresume.MappedFiles = append(transfer.Fastresume.MappedFiles, fileHelpers.Normalize(helpers.HandleCesu8(filePath), transfer.Opts.PathSeparator))
				}
				// and then doing remap if resumeItem contain target field
				if maxIndex := transfer.FindHighestIndexOfMappedFiles(); maxIndex >= 0 {
					for _, paths := range transfer.ResumeItem.Targets {
						index := paths[0].(int64)
						var pathParts []string
						if fileHelpers.IsAbs(helpers.HandleCesu8(paths[1].(string))) {
							pathParts = []string{fileHelpers.Normalize(helpers.HandleCesu8(paths[1].(string)), transfer.Opts.PathSeparator)}
						} else {
							pathParts = make([]string, len(paths)-1, len(paths)-1)
							for num, part := range paths[1:] {
								pathParts[num] = helpers.HandleCesu8(part.(string))
							}
						}
						transfer.Fastresume.MappedFiles[index] = fileHelpers.Join(pathParts, transfer.Opts.PathSeparator)
					}
				}
				transfer.Fastresume.QbtSavePath = fileHelpers.Normalize(helpers.HandleCesu8(transfer.ResumeItem.Path), "/")
			}
		} else {
			transfer.Fastresume.QBtContentLayout = "Original" // utorrent\bittorrent don't support create subfolders for torrents with single file
			if lastPathName != torrentNameOriginal {
				//it means that we have renamed path and targets item, and should have mapped files
				transfer.Fastresume.MappedFiles = []string{lastPathName}
			}
			transfer.Fastresume.QbtSavePath = fileHelpers.CutLastPath(helpers.HandleCesu8(transfer.ResumeItem.Path), `/`)
			if string(transfer.Fastresume.QbtSavePath[len(transfer.Fastresume.QbtSavePath)-1]) != `/` {
				transfer.Fastresume.QbtSavePath += `/`
			}
		}

		// transform windows prohibited symbols like libtorrent or utorrent do
		if transfer.Opts.PathSeparator == `\` && transfer.Fastresume.MappedFiles != nil {
			for index, mappedFile := range transfer.Fastresume.MappedFiles {
				if fileHelpers.IsAbs(mappedFile) {
					transfer.Fastresume.MappedFiles[index] = mappedFile[:2] + prohibitedSymbols.ReplaceAllString(mappedFile[2:], `_`)
				} else {
					transfer.Fastresume.MappedFiles[index] = prohibitedSymbols.ReplaceAllString(mappedFile, `_`)
				}
			}
		}
	}

	for _, pattern := range transfer.Replace {
		transfer.Fastresume.QbtSavePath = strings.ReplaceAll(transfer.Fastresume.QbtSavePath, pattern.From, pattern.To)
		// replace mapped files if them are absolute paths
		for mapIndex, mapPath := range transfer.Fastresume.MappedFiles {
			if fileHelpers.IsAbs(mapPath) {
				transfer.Fastresume.MappedFiles[mapIndex] = strings.ReplaceAll(mapPath, pattern.From, pattern.To)
			}
		}
	}

	transfer.Fastresume.SavePath = fileHelpers.Normalize(transfer.Fastresume.QbtSavePath, transfer.Opts.PathSeparator)
	if transfer.Fastresume.QBtContentLayout == "Original" && !transfer.Magnet {
		if string(transfer.Fastresume.SavePath[len(transfer.Fastresume.SavePath)-1]) != transfer.Opts.PathSeparator {
			transfer.Fastresume.SavePath += transfer.Opts.PathSeparator
		}
	}
}

// FindHighestIndexOfMappedFiles just helper for creating mappedfiles
func (transfer *TransferStructure) FindHighestIndexOfMappedFiles() int64 {
	if resumeItem := transfer.ResumeItem; resumeItem.Targets != nil {
		lastElem := resumeItem.Targets[len(resumeItem.Targets)-1] // it must be like []interface{0, "path"}
		return lastElem[0].(int64)
	}
	return -1
}

func CreateReplaces(replaces []string) []*replace.Replace {
	var r []*replace.Replace
	for _, str := range replaces {
		patterns := strings.Split(str, ",")
		r = append(r, &replace.Replace{
			From: patterns[0],
			To:   patterns[1],
		})
	}
	return r
}
