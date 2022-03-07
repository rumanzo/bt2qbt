package test

import (
	"github.com/rumanzo/bt2qbt/internal/libtorrent"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/qBittorrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"testing"
)

var newstructure = libtorrent.NewTorrentStructure{
	Fastresume: qBittorrentStructures.QBittorrentFastresume{
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
		LibTorrentVersion:   "1.2.5.0",
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
	TorrentFile: &torrentStructures.Torrent{},
}

func TestPath(t *testing.T) {
	err := helpers.DecodeTorrentFile("./data/testfileset.torrent", newstructure.TorrentFile)
	if err != nil {
		t.Fatalf("Can't decode torrent file with error: %v", err)
	}
}
