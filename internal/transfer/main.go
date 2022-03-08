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
)

func HandleResumeItem(key string, resumeItem *utorrentStructs.ResumeItem, opts *options.Opts, chans *Channels, wg *sync.WaitGroup) error {

	//panic recover
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

	// preparing structures for work with
	var err error
	newStructure := libtorrent.CreateEmptyNewTorrentStructure()
	newStructure.WithoutTags = opts.WithoutTags
	newStructure.WithoutLabels = opts.WithoutLabels
	newStructure.ResumeItem = resumeItem
	for _, str := range opts.Replaces {
		patterns := strings.Split(str, ",")
		newStructure.Replace = append(newStructure.Replace, replace.Replace{
			From: patterns[0],
			To:   patterns[1],
		})
	}

	handleTorrentFilePath(newStructure, key, opts)

	err = findTorrentFile(newStructure, opts.SearchPaths)
	if err != nil {
		chans.ErrChannel <- err.Error()
		return err
	}

	err = helpers.DecodeTorrentFile(newStructure.TorrentFilePath, newStructure.TorrentFile)
	if err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't decode torrent file %v for %v", newStructure.TorrentFilePath, key)
		return err
	}

	newStructure.HandleStructures()

	newbasename := newStructure.GetHash()
	if err = helpers.EncodeTorrentFile(filepath.Join(opts.QBitDir, newbasename+".fastresume"), &newStructure); err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't create qBittorrent fastresume file %v", opts.QBitDir+newbasename+".fastresume")
		return err
	}
	if err = helpers.CopyFile(newStructure.TorrentFilePath, filepath.Join(opts.QBitDir, newbasename+".torrent")); err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't create qBittorrent torrent file %v", filepath.Join(opts.QBitDir, newbasename+".torrent"))
		return err
	}
	chans.ComChannel <- fmt.Sprintf("Sucessfully imported %v", key)
	return nil
}

func HandleResumeItems(opts *options.Opts, resumeItems map[string]*utorrentStructs.ResumeItem) {
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
			go HandleResumeItem(key, resumeItem, opts, &chans, &wg)
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

// check if resume key is absolute path. It means that we should search torrent file using this absolute path
// notice that torrent file name always known
func handleTorrentFilePath(newStructure libtorrent.NewTorrentStructure, key string, opts *options.Opts) {
	if isAbs, _ := regexp.MatchString(`^([A-Za-z]:)?\\\\?`, key); isAbs == true {
		if runtime.GOOS == "windows" {
			newStructure.TorrentFilePath = key
			newStructure.TorrentFileName = filepath.Base(key)
		} else { // for unix system find in search paths, we just get basename of torrent file
			newStructure.TorrentFileName = filepath.Base(key)
		}
	} else {
		newStructure.TorrentFilePath = filepath.Join(opts.BitDir, key) // additional search required
		newStructure.TorrentFileName = key
	}
}

// if we can find torrent file, we start check another locations from options search paths
func findTorrentFile(newStructure libtorrent.NewTorrentStructure, searchPaths []string) error {
	if _, err := os.Stat(newStructure.TorrentFilePath); os.IsNotExist(err) {
		for _, searchPath := range searchPaths {
			if _, err = os.Stat(filepath.Join(searchPath, newStructure.TorrentFileName)); err == nil {
				newStructure.TorrentFilePath = filepath.Join(searchPath, newStructure.TorrentFileName)
				return nil
			}
		}
	}
	return fmt.Errorf("Can't locate torrent file %v", newStructure.TorrentFileName)
}
