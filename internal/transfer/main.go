package transfer

import (
	"fmt"
	"github.com/rumanzo/bt2qbt/internal/libtorrent"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/internal/replace"
	"github.com/rumanzo/bt2qbt/pkg/fileHelpers"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"log"
	"os"
	"path/filepath"
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

	HandleTorrentFilePath(&newStructure, key, opts)

	err = FindTorrentFile(&newStructure, opts.SearchPaths)
	if err != nil {
		chans.ErrChannel <- err.Error()
		return err
	}

	// struct for work with
	err = helpers.DecodeTorrentFile(newStructure.TorrentFilePath, newStructure.TorrentFile)
	if err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't decode torrent file %v for %v", newStructure.TorrentFilePath, key)
		return err
	}

	// because hash of info very important it will be better to use interface for get hash
	err = helpers.DecodeTorrentFile(newStructure.TorrentFilePath, &newStructure.TorrentFileRaw)
	if err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't decode torrent file %v for %v", newStructure.TorrentFilePath, key)
		return err
	}

	newStructure.HandleStructures()

	newBaseName := newStructure.GetHash()
	if err = helpers.EncodeTorrentFile(filepath.Join(opts.QBitDir, newBaseName+".fastresume"), newStructure.Fastresume); err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't create qBittorrent fastresume file %v. With error: %v", filepath.Join(opts.QBitDir, newBaseName+".fastresume"), err)
		return err
	}
	if err = helpers.CopyFile(newStructure.TorrentFilePath, filepath.Join(opts.QBitDir, newBaseName+".torrent")); err != nil {
		chans.ErrChannel <- fmt.Sprintf("Can't create qBittorrent torrent file %v", filepath.Join(opts.QBitDir, newBaseName+".torrent"))
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
	numJob := 1
	var newTags []string
	var wg sync.WaitGroup

	positionNum := 0

	for key, resumeItem := range resumeItems {
		positionNum++
		if opts.WithoutTags == false {
			if resumeItem.Labels != nil {
				for _, label := range resumeItem.Labels {
					if exists, tag := helpers.CheckExists(helpers.ASCIIConvert(label), newTags); !exists {
						newTags = append(newTags, tag)
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
		fmt.Printf("%v/%v %v \n", numJob, totalJobs, message)
		numJob++
	}
	var wasErrors bool
	for message := range chans.ErrChannel {
		fmt.Printf("%v/%v %v \n", numJob, totalJobs, message)
		wasErrors = true
		numJob++
	}
	if opts.WithoutTags == false {
		ProcessLabels(opts, newTags)
	}
	fmt.Println()
	log.Println("Ended")
	if wasErrors {
		log.Println("Not all torrents was processed")
	}
}

// check if resume key is absolute path. It means that we should search torrent file using this absolute path
// notice that torrent file name always known
func HandleTorrentFilePath(newStructure *libtorrent.NewTorrentStructure, key string, opts *options.Opts) {
	if fileHelpers.IsAbs(key) {
		newStructure.TorrentFilePath = key
		newStructure.TorrentFileName = fileHelpers.Base(key)
	} else {
		newStructure.TorrentFilePath = filepath.Join(opts.BitDir, key) // additional search required
		newStructure.TorrentFileName = key
	}
}

// if we can find torrent file, we start check another locations from options search paths
func FindTorrentFile(newStructure *libtorrent.NewTorrentStructure, searchPaths []string) error {
	if _, err := os.Stat(newStructure.TorrentFilePath); os.IsNotExist(err) {
		for _, searchPath := range searchPaths {
			// normalize path with os.PathSeparator, because file can be on share, for example
			fullPath := fileHelpers.Join([]string{searchPath, newStructure.TorrentFileName}, string(os.PathSeparator))
			if _, err = os.Stat(fullPath); err == nil {
				newStructure.TorrentFilePath = fullPath
				return nil
			}
		}
		// return error only if we didn't find anything
		return fmt.Errorf("Can't locate torrent file %v", newStructure.TorrentFileName)
	}
	return nil
}
