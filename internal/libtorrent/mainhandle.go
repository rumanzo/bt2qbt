package libtorrent

import (
	"time"
)

func (newStructure *NewTorrentStructure) HandleStructures() {

	if ok := newStructure.ResumeItem.Targets; ok != nil {
		for _, entry := range newStructure.ResumeItem.Targets {
			newStructure.Targets[entry[0].(int64)] = entry[1].(string)
		}
	}

	// if torrent name was renamed, add modified name
	newStructure.HandleCaption()
	newStructure.Fastresume.ActiveTime = newStructure.ResumeItem.Runtime
	newStructure.Fastresume.AddedTime = newStructure.ResumeItem.AddedOn
	newStructure.Fastresume.CompletedTime = newStructure.ResumeItem.CompletedOn
	newStructure.Fastresume.Info = newStructure.TorrentFile.Info

	// todo
	//torinfo, _ := bencode.EncodeString(newStructure.TorrentFile.Info)
	//h := sha1.New()
	//io.WriteString(h, torinfo)
	//h.Sum(nil)
	//
	//newStructure.Fastresume.InfoHash = torinfo

	newStructure.Fastresume.SeedingTime = newStructure.ResumeItem.Runtime
	newStructure.HandleState()
	newStructure.Fastresume.FinishedTime = int64(time.Since(time.Unix(newStructure.ResumeItem.CompletedOn, 0)).Minutes())

	newStructure.HandleTotalDownloaded()
	newStructure.Fastresume.TotalUploaded = newStructure.ResumeItem.Uploaded
	newStructure.Fastresume.UploadRateLimit = newStructure.ResumeItem.UpSpeed
	newStructure.HandleTags()
	newStructure.HandleLabels()

	newStructure.GetTrackers(newStructure.ResumeItem.Trackers)
	newStructure.PrioConvert(newStructure.ResumeItem.Prio)

	// https://libtorrent.org/manual-ref.html#fast-resume
	newStructure.PieceLenght = newStructure.TorrentFile.Info.PieceLength

	/*
		pieces maps to a string whose length is a multiple of 20. It is to be subdivided into strings of length 20,
		each of which is the SHA1 hash of the piece at the corresponding index.
		http://www.bittorrent.org/beps/bep_0003.html
	*/
	newStructure.NumPieces = int64(len(newStructure.TorrentFile.Info.Pieces)) / 20

	newStructure.HandleCompleted()
	newStructure.HandleSizes()
	newStructure.HandleSavePaths()
	newStructure.HandlePieces()
}
