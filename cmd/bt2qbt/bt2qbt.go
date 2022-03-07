package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-ini/ini"
	"github.com/rumanzo/bt2qbt/internal/libtorrent"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/internal/replace"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/qBittorrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"github.com/zeebo/bencode"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Channels struct {
	comChannel     chan string
	errChannel     chan string
	boundedChannel chan bool
}

func logic(key string, resumeItem *utorrentStructs.ResumeItem, opts *options.Opts, chans *Channels, position int, wg *sync.WaitGroup) error {
	defer wg.Done()
	defer func() {
		<-chans.boundedChannel
	}()
	defer func() {
		if r := recover(); r != nil {
			chans.errChannel <- fmt.Sprintf(
				"Panic while processing torrent %v:\n======\nReason: %v.\nText panic:\n%v\n======",
				key, r, string(debug.Stack()))
		}
	}()
	var err error
	newstructure := libtorrent.NewTorrentStructure{
		Fastresume: qBittorrentStructures.QBittorrentFastresume{
			ActiveTime:          0,
			AddedTime:           0,
			Allocation:          "sparse",
			AutoManaged:         0,
			CompletedTime:       0,
			DownloadRateLimit:   -1,
			FileFormat:          "libtorrent resume file",
			FileVersion:         1,
			FilePriority:        []int64{},
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
		TorrentFile:   &torrentStructures.Torrent{},
		WithoutLabels: opts.WithoutLabels,
		WithoutTags:   opts.WithoutTags,
		Separator:     opts.PathSeparator,
		Targets:       map[int64]string{},
	}

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
		chans.errChannel <- fmt.Sprintf("Can't find torrent file %v for %v", newstructure.TorrentFilePath, key)
		return err
	CONTINUE:
	}
	err = helpers.DecodeTorrentFile(newstructure.TorrentFilePath, newstructure.TorrentFile)
	if err != nil {
		chans.errChannel <- fmt.Sprintf("Can't decode torrent file %v for %v", newstructure.TorrentFilePath, key)
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

	if ok := resumeItem.Targets; ok != nil {
		for _, entry := range resumeItem.Targets {
			newstructure.Targets[entry[0].(int64)] = entry[1].(string)
		}
	}

	newstructure.Path = resumeItem.Path

	// if torrent name was renamed, add modified name
	if resumeItem.Caption != "" {
		newstructure.Fastresume.QbtName = resumeItem.Caption
	}
	newstructure.Fastresume.ActiveTime = resumeItem.Runtime
	newstructure.Fastresume.AddedTime = resumeItem.AddedOn
	newstructure.Fastresume.CompletedTime = resumeItem.CompletedOn
	//newstructure.Fastresume.InfoHash = value["info"].(string) //todo
	newstructure.Fastresume.SeedingTime = resumeItem.Runtime
	if resumeItem.Started == 0 {
		newstructure.Fastresume.Paused = 1
		newstructure.Fastresume.AutoManaged = 0
	} else {
		newstructure.Fastresume.Paused = 0
		newstructure.Fastresume.AutoManaged = 1
	}

	newstructure.Fastresume.FinishedTime = int64(time.Since(time.Unix(resumeItem.CompletedOn, 0)).Minutes())
	if resumeItem.CompletedOn == 0 {
		newstructure.Fastresume.TotalDownloaded = 0
	} else {
		newstructure.Fastresume.TotalDownloaded = resumeItem.Downloaded
	}
	newstructure.Fastresume.TotalUploaded = resumeItem.Uploaded
	newstructure.Fastresume.UploadRateLimit = resumeItem.UpSpeed
	newstructure.IfTags(resumeItem.Labels)
	if resumeItem.Label != "" {
		newstructure.IfLabel(resumeItem.Label)
	} else {
		newstructure.IfLabel("")
	}
	newstructure.GetTrackers(resumeItem.Trackers)
	newstructure.PrioConvert(resumeItem.Prio)

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
		chans.errChannel <- fmt.Sprintf("Can't create qBittorrent fastresume file %v", opts.QBitDir+newbasename+".fastresume")
		return err
	}
	if err = helpers.CopyFile(newstructure.TorrentFilePath, opts.QBitDir+newbasename+".torrent"); err != nil {
		chans.errChannel <- fmt.Sprintf("Can't create qBittorrent torrent file %v", opts.QBitDir+newbasename+".torrent")
		return err
	}
	chans.comChannel <- fmt.Sprintf("Sucessfully imported %v", key)
	return nil
}

func main() {
	opts := options.MakeOpts()

	resumeFilePath := path.Join(opts.BitDir, "resume.dat")
	if _, err := os.Stat(resumeFilePath); os.IsNotExist(err) {
		log.Println("Can't find uTorrent\\Bittorrent resume file")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	resumeFile := map[string]interface{}{}
	err := helpers.DecodeTorrentFile(resumeFilePath, resumeFile)
	if err != nil {
		log.Println("Can't decode uTorrent\\Bittorrent resume file")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	totalJobs := len(resumeFile)
	chans := Channels{comChannel: make(chan string, totalJobs),
		errChannel:     make(chan string, totalJobs),
		boundedChannel: make(chan bool, runtime.GOMAXPROCS(0)*2)}

	color.Green("It will be performed processing from directory %v to directory %v\n", opts.BitDir, opts.QBitDir)
	color.HiRed("Check that the qBittorrent is turned off and the directory %v and config %v is backed up.\n",
		opts.QBitDir, opts.Config)
	color.HiRed("Check that you previously disable option \"Append .!ut/.!bt to incomplete files\" in preferences of uTorrent/Bittorrent \n\n")
	fmt.Println("Press Enter to start")
	fmt.Scanln()
	log.Println("Started")

	transferTorrents(chans, opts, resumeFile, totalJobs)

	fmt.Println("\nPress Enter to exit")
	fmt.Scanln()

}

func transferTorrents(chans Channels, opts *options.Opts, resumeFile map[string]interface{}, totalJobs int) {
	numjob := 1
	var oldtags string
	var newtags []string
	var wg sync.WaitGroup

	positionnum := 0

	// hate utorrent for heterogeneous resume.dat scheme
	delete(resumeFile, ".fileguard")
	delete(resumeFile, "rec")
	b, _ := bencode.EncodeBytes(resumeFile)
	resumeItems := map[string]*utorrentStructs.ResumeItem{}
	bencode.DecodeBytes(b, &resumeItems)

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
			chans.boundedChannel <- true
			go logic(key, resumeItem, opts, &chans, positionnum, &wg)
		} else {
			totalJobs--
		}
	}
	go func() {
		wg.Wait()
		close(chans.comChannel)
		close(chans.errChannel)
	}()
	for message := range chans.comChannel {
		fmt.Printf("%v/%v %v \n", numjob, totalJobs, message)
		numjob++
	}
	var waserrors bool
	for message := range chans.errChannel {
		fmt.Printf("%v/%v %v \n", numjob, totalJobs, message)
		waserrors = true
		numjob++
	}
	if opts.WithoutTags == false {
		cfg, err := ini.Load(opts.Config)
		ini.PrettyFormat = false
		ini.PrettySection = false
		if err != nil {
			fmt.Println("Can not read qBittorrent config file. Try to specify the path explicitly or do not import tags")
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
		if _, err := cfg.GetSection("BitTorrent"); err != nil {
			cfg.NewSection("BitTorrent")

			//Dirty hack for section order. Sorry
			kv := cfg.Section("Network").KeysHash()
			cfg.DeleteSection("Network")
			cfg.NewSection("Network")
			for key, value := range kv {
				cfg.Section("Network").NewKey(key, value)
			}
			//End of dirty hack
		}
		if cfg.Section("BitTorrent").HasKey("Session\\Tags") {
			oldtags = cfg.Section("BitTorrent").Key("Session\\Tags").String()
			for _, tag := range strings.Split(oldtags, ", ") {
				if exists, t := helpers.CheckExists(tag, newtags); !exists {
					newtags = append(newtags, t)
				}
			}
			cfg.Section("BitTorrent").Key("Session\\Tags").SetValue(strings.Join(newtags, ", "))
		} else {
			cfg.Section("BitTorrent").NewKey("Session\\Tags", strings.Join(newtags, ", "))
		}
		cfg.SaveTo(opts.Config)
	}
	fmt.Println()
	log.Println("Ended")
	if waserrors {
		log.Println("Not all torrents was processed")
	}
}
