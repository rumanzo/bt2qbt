package test

import (
	"github.com/rumanzo/bt2qbt/internal/libtorrent"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"testing"
)

func TestPath(t *testing.T) {
	nts := libtorrent.CreateEmptyNewTorrentStructure()
	err := helpers.DecodeTorrentFile("./data/testfileset.torrent", nts)
	if err != nil {
		t.Fatalf("Can't decode torrent file with error: %v", err)
	}
}
