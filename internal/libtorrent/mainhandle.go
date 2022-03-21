package libtorrent

import (
	"time"
)

func (transfer *TransferStructure) HandleStructures() {

	if ok := transfer.ResumeItem.Targets; ok != nil {
		for _, entry := range transfer.ResumeItem.Targets {
			transfer.Targets[entry[0].(int64)] = entry[1].(string)
		}
	}

	// if torrent name was renamed, add modified name
	transfer.HandleCaption()
	transfer.Fastresume.ActiveTime = transfer.ResumeItem.Runtime
	transfer.Fastresume.AddedTime = transfer.ResumeItem.AddedOn
	transfer.Fastresume.CompletedTime = transfer.ResumeItem.CompletedOn
	transfer.Fastresume.Info = transfer.TorrentFile.Info
	transfer.Fastresume.InfoHash = transfer.ResumeItem.Info
	transfer.Fastresume.SeedingTime = transfer.ResumeItem.Runtime
	transfer.HandleState()
	transfer.Fastresume.FinishedTime = int64(time.Since(time.Unix(transfer.ResumeItem.CompletedOn, 0)).Minutes())

	transfer.HandleTotalDownloaded()
	transfer.Fastresume.TotalUploaded = transfer.ResumeItem.Uploaded
	transfer.Fastresume.UploadRateLimit = transfer.ResumeItem.UpSpeed
	transfer.HandleTags()
	transfer.HandleLabels()

	transfer.GetTrackers(transfer.ResumeItem.Trackers)
	transfer.HandlePriority() // important handle priorities before handling pieces

	/*
		pieces maps to a string whose length is a multiple of 20. It is to be subdivided into strings of length 20,
		each of which is the SHA1 hash of the piece at the corresponding index.
		http://www.bittorrent.org/beps/bep_0003.html
	*/
	transfer.NumPieces = int64(len(transfer.TorrentFile.Info.Pieces)) / 20

	transfer.HandleCompleted()
	transfer.HandleSavePaths()
	transfer.HandlePieces()
}
