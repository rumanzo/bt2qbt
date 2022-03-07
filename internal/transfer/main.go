package transfer

import (
	"fmt"
	"github.com/rumanzo/bt2qbt/internal/libtorrent"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/internal/replace"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

func ProcessResumeItem(key string, resumeItem *utorrentStructs.ResumeItem, opts *options.Opts, chans *Channels, wg *sync.WaitGroup) error {

	defer wg.Done()
	defer func() {
		<-chans.BoundedChannel
	}()
	defer func() {
		if r := recover(); r != nil {
			chans.ErrChannel <- fmt.Sprintf(
				"Panic while processing torrent %v:\n======\nReason: %v.\nText panic:\n%v\n======",
				key, r, string(debug.Stack()))
		}
	}()
	var err error
	newstructure := libtorrent.CreateEmptyNewTorrentStructure()
	newstructure.ResumeItem = resumeItem

	if isAbs, _ := regexp.MatchString(`^([A-Za-z]:)?\\`, key); isAbs == true {
		if runtime.GOOS == "windows" {
			newstructure.TorrentFilePath = key
		} else { // for unix system find in search paths
			pathparts := strings.Split(key, "\\")
			newstructure.TorrentFilePath = pathparts[len(pathparts)-1]
		}
	} else {
		newstructure.TorrentFilePath = filepath.Join(opts.BitDir, key) // additional search required
	}
	if _, err = os.Stat(newstructure.TorrentFilePath); os.IsNotExist(err) {
		for _, searchPath := range opts.SearchPaths {
			if _, err = os.Stat(searchPath + newstructure.TorrentFilePath); err == nil {
				newstructure.TorrentFilePath = searchPath + newstructure.TorrentFilePath
				goto CONTINUE
			}
		}
		chans.ErrChannel <- fmt.Sprintf("Can't find torrent file %v for %v", newstructure.TorrentFilePath, key)
		return err
	CONTINUE:
	}
	err = helpers.DecodeTorrentFile(newstructure.TorrentFilePath, newstructure.TorrentFile)
	if err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't decode torrent file %v for %v", newstructure.TorrentFilePath, key)
		return err
	}

	for _, str := range opts.Replaces {
		patterns := strings.Split(str, ",")
		newstructure.Replace = append(newstructure.Replace, replace.Replace{
			From: patterns[0],
			To:   patterns[1],
		})
	}

	if len(newstructure.TorrentFile.Info.Files) > 0 {
		newstructure.HasFiles = true
	} else {
		newstructure.HasFiles = false
	}

	if ok := newstructure.ResumeItem.Targets; ok != nil {
		for _, entry := range newstructure.ResumeItem.Targets {
			newstructure.Targets[entry[0].(int64)] = entry[1].(string)
		}
	}

	newstructure.Path = newstructure.ResumeItem.Path

	// if torrent name was renamed, add modified name
	if newstructure.ResumeItem.Caption != "" {
		newstructure.Fastresume.QbtName = newstructure.ResumeItem.Caption
	}
	newstructure.Fastresume.ActiveTime = newstructure.ResumeItem.Runtime
	newstructure.Fastresume.AddedTime = newstructure.ResumeItem.AddedOn
	newstructure.Fastresume.CompletedTime = newstructure.ResumeItem.CompletedOn
	//newstructure.Fastresume.InfoHash = value["info"].(string) //todo
	newstructure.Fastresume.SeedingTime = newstructure.ResumeItem.Runtime
	if newstructure.ResumeItem.Started == 0 {
		newstructure.Fastresume.Paused = 1
		newstructure.Fastresume.AutoManaged = 0
	} else {
		newstructure.Fastresume.Paused = 0
		newstructure.Fastresume.AutoManaged = 1
	}

	newstructure.Fastresume.FinishedTime = int64(time.Since(time.Unix(newstructure.ResumeItem.CompletedOn, 0)).Minutes())
	if newstructure.ResumeItem.CompletedOn == 0 {
		newstructure.Fastresume.TotalDownloaded = 0
	} else {
		newstructure.Fastresume.TotalDownloaded = newstructure.ResumeItem.Downloaded
	}
	newstructure.Fastresume.TotalUploaded = newstructure.ResumeItem.Uploaded
	newstructure.Fastresume.UploadRateLimit = newstructure.ResumeItem.UpSpeed
	newstructure.IfTags(newstructure.ResumeItem.Labels)
	if newstructure.ResumeItem.Label != "" {
		newstructure.IfLabel(newstructure.ResumeItem.Label)
	} else {
		newstructure.IfLabel("")
	}
	newstructure.GetTrackers(newstructure.ResumeItem.Trackers)
	newstructure.PrioConvert(newstructure.ResumeItem.Prio)

	// https://libtorrent.org/manual-ref.html#fast-resume
	newstructure.PieceLenght = newstructure.TorrentFile.Info.PieceLength

	/*
		pieces maps to a string whose length is a multiple of 20. It is to be subdivided into strings of length 20,
		each of which is the SHA1 hash of the piece at the corresponding index.
		http://www.bittorrent.org/beps/bep_0003.html
	*/
	newstructure.NumPieces = int64(len(newstructure.TorrentFile.Info.Pieces)) / 20
	newstructure.FillMissing()
	newbasename := newstructure.GetHash()

	if err = libtorrent.EncodeTorrentFile(opts.QBitDir+newbasename+".fastresume", &newstructure); err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't create qBittorrent fastresume file %v", opts.QBitDir+newbasename+".fastresume")
		return err
	}
	if err = helpers.CopyFile(newstructure.TorrentFilePath, opts.QBitDir+newbasename+".torrent"); err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't create qBittorrent torrent file %v", opts.QBitDir+newbasename+".torrent")
		return err
	}
	chans.ComChannel <- fmt.Sprintf("Sucessfully imported %v", key)
	return nil
}

func TransferTorrents(opts *options.Opts, resumeItems map[string]*utorrentStructs.ResumeItem) {
	totalJobs := len(resumeItems)
	chans := Channels{ComChannel: make(chan string, totalJobs),
		ErrChannel:     make(chan string, totalJobs),
		BoundedChannel: make(chan bool, runtime.GOMAXPROCS(0)*2)}
	numjob := 1
	var newtags []string
	var wg sync.WaitGroup

	positionnum := 0

	for key, resumeItem := range resumeItems {
		positionnum++
		if opts.WithoutTags == false {
			if resumeItem.Labels != nil {
				for _, label := range resumeItem.Labels {
					if exists, tag := helpers.CheckExists(helpers.ASCIIConvert(label), newtags); !exists {
						newtags = append(newtags, tag)
					}
				}
			}
			wg.Add(1)
			chans.BoundedChannel <- true
			go ProcessResumeItem(key, resumeItem, opts, &chans, &wg)
		} else {
			totalJobs--
		}
	}
	go func() {
		wg.Wait()
		close(chans.ComChannel)
		close(chans.ErrChannel)
	}()
	for message := range chans.ComChannel {
		fmt.Printf("%v/%v %v \n", numjob, totalJobs, message)
		numjob++
	}
	var waserrors bool
	for message := range chans.ErrChannel {
		fmt.Printf("%v/%v %v \n", numjob, totalJobs, message)
		waserrors = true
		numjob++
	}
	if opts.WithoutTags == false {
		ProcessLabels(opts, newtags)
	}
	fmt.Println()
	log.Println("Ended")
	if waserrors {
		log.Println("Not all torrents was processed")
	}
}
