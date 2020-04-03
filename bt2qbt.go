package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-ini/ini"
	goflags "github.com/jessevdk/go-flags"
	"github.com/rumanzo/bt2qbt/libtorrent"
	"github.com/rumanzo/bt2qbt/replace"
	"github.com/zeebo/bencode"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Flags struct {
	BitDir        string   `short:"s" long:"source" description:"Source directory that contains resume.dat and torrents files"`
	QBitDir       string   `short:"d" long:"destination" description:"Destination directory BT_backup (as default)"`
	Config        string   `short:"c" long:"config" description:"Source directory that contains resume.dat and torrents files"`
	WithoutLabels bool     `long:"without-labels" description:"Do not export/import labels"`
	WithoutTags   bool     `long:"without-tags" description:"Do not export/import tags"`
	Replace       []string `short:"r" long:"replace" description:"Replace paths.\n	Delimiter for from/to - ,\n	Example: \"D:\\films,/home/user/films;\\,/\"\n	If you use path separator different from you system, declare it mannually"`
}

type Channels struct {
	comChannel     chan string
	errChannel     chan string
	boundedChannel chan bool
}

func encodetorrentfile(path string, newstructure *libtorrent.NewTorrentStructure) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Create(path)
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	bufferedWriter := bufio.NewWriter(file)
	enc := bencode.NewEncoder(bufferedWriter)
	if err := enc.Encode(newstructure); err != nil {
		return err
	}
	bufferedWriter.Flush()
	return nil
}

func ASCIIconvert(s string) string {
	var buffer bytes.Buffer
	for _, c := range s {
		if c > 127 {
			buffer.WriteString(`\x` + strconv.FormatUint(uint64(c), 16))
		} else {
			buffer.WriteString(string(c))
		}
	}
	return buffer.String()
}

func checknotexists(s string, tags []string) (bool, string) {
	for _, value := range tags {
		if value == s {
			return false, s
		}
	}
	return true, s
}

func decodetorrentfile(path string) (map[string]interface{}, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var torrent map[string]interface{}
	if err := bencode.DecodeBytes(dat, &torrent); err != nil {
		return nil, err
	}
	return torrent, nil
}

func copyfile(src string, dst string) error {
	originalFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer originalFile.Close()
	newFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer newFile.Close()
	if _, err := io.Copy(newFile, originalFile); err != nil {
		return err
	}
	if err := newFile.Sync(); err != nil {
		return err
	}
	return nil
}

func logic(key string, value map[string]interface{}, flags *Flags, chans *Channels, position int, wg *sync.WaitGroup) error {
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
	newstructure := libtorrent.NewTorrentStructure{ActiveTime: 0, AddedTime: 0, AnnounceToDht: 0, AnnounceToLsd: 0,
		AnnounceToTrackers: 0, AutoManaged: 0, CompletedTime: 0, DownloadRateLimit: -1,
		FileFormat: "libtorrent resume file", FileVersion: 1, FinishedTime: 0, LastSeenComplete: 0,
		LibTorrentVersion: "1.1.6.0", MaxConnections: 100, MaxUploads: 100, NumDownloaded: 0, NumIncomplete: 0,
		QbtQueuePosition: 1, QbtRatioLimit: -2000, QbtSeedStatus: 1, QbtSeedingTimeLimit: -2, QbttempPathDisabled: 0,
		SeedMode: 0, SeedingTime: 0, SequentialDownload: 0, SuperSeeding: 0, TotalDownloaded: 0, TotalUploaded: 0,
		UploadRateLimit: 0, QbtName: "", WithoutLabels: flags.WithoutLabels, WithoutTags: flags.WithoutTags}
	if ok := filepath.IsAbs(key); ok {
		newstructure.TorrentFilePath = key
	} else {
		newstructure.TorrentFilePath = flags.BitDir + key
	}
	if _, err = os.Stat(newstructure.TorrentFilePath); os.IsNotExist(err) {
		chans.errChannel <- fmt.Sprintf("Can't find torrent file %v for %v", newstructure.TorrentFilePath, key)
		return err
	}
	newstructure.TorrentFile, err = decodetorrentfile(newstructure.TorrentFilePath)
	if err != nil {
		chans.errChannel <- fmt.Sprintf("Can't decode torrent file %v for %v", newstructure.TorrentFilePath, key)
		return err
	}

	if len(flags.Replace) != 0 {
		for _, str := range flags.Replace {
			patterns := strings.Split(str, ",")
			newstructure.Replace = append(newstructure.Replace, replace.Replace{
				From: patterns[0],
				To:   patterns[1],
			})
		}
	}

	if _, ok := newstructure.TorrentFile["info"].(map[string]interface{})["files"]; ok {
		newstructure.HasFiles = true
	} else {
		newstructure.HasFiles = false
	}
	if value["path"].(string)[len(value["path"].(string))-1] == os.PathSeparator {
		newstructure.Path = value["path"].(string)[:len(value["path"].(string))-1]
	} else {
		newstructure.Path = value["path"].(string)
	}
	newstructure.ActiveTime = value["runtime"].(int64)
	newstructure.AddedTime = value["added_on"].(int64)
	newstructure.CompletedTime = value["completed_on"].(int64)
	newstructure.InfoHash = value["info"].(string)
	newstructure.SeedingTime = value["runtime"].(int64)
	newstructure.QbtQueuePosition = position
	newstructure.Started(value["started"].(int64))
	newstructure.FinishedTime = int64(time.Since(time.Unix(value["completed_on"].(int64), 0)).Minutes())
	newstructure.TotalDownloaded = value["downloaded"].(int64)
	newstructure.TotalUploaded = value["uploaded"].(int64)
	newstructure.UploadRateLimit = value["upspeed"].(int64)
	newstructure.IfTags(value["labels"])
	if value["label"] != nil {
		newstructure.IfLabel(value["label"].(string))
	} else {
		newstructure.IfLabel("")
	}
	newstructure.GetTrackers(value["trackers"])
	newstructure.PrioConvert(value["prio"].(string))

	// https://libtorrent.org/manual-ref.html#fast-resume
	newstructure.BlockPerPiece = newstructure.TorrentFile["info"].(map[string]interface{})["piece length"].(int64) / 16 / 1024
	newstructure.PieceLenght = newstructure.TorrentFile["info"].(map[string]interface{})["piece length"].(int64)

	/*
		pieces maps to a string whose length is a multiple of 20. It is to be subdivided into strings of length 20,
		each of which is the SHA1 hash of the piece at the corresponding index.
		http://www.bittorrent.org/beps/bep_0003.html
	*/
	newstructure.NumPieces = int64(len(newstructure.TorrentFile["info"].(map[string]interface{})["pieces"].(string))) / 20
	newstructure.FillMissing()
	newbasename := newstructure.GetHash()

	if err = encodetorrentfile(flags.QBitDir+newbasename+".fastresume", &newstructure); err != nil {
		chans.errChannel <- fmt.Sprintf("Can't create qBittorrent fastresume file %v", flags.QBitDir+newbasename+".fastresume")
		return err
	}
	if err = copyfile(newstructure.TorrentFilePath, flags.QBitDir+newbasename+".torrent"); err != nil {
		chans.errChannel <- fmt.Sprintf("Can't create qBittorrent torrent file %v", flags.QBitDir+newbasename+".torrent")
		return err
	}
	chans.comChannel <- fmt.Sprintf("Sucessfully imported %v", key)
	return nil
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flags := Flags{}
	sep := string(os.PathSeparator)
	switch OS := runtime.GOOS; OS {
	case "windows":
		flags.BitDir = os.Getenv("APPDATA") + sep + "uTorrent" + sep
		flags.Config = os.Getenv("APPDATA") + sep + "qBittorrent" + sep + "qBittorrent.ini"
		flags.QBitDir = os.Getenv("LOCALAPPDATA") + sep + "qBittorrent" + sep + "BT_backup" + sep
	case "linux":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		flags.BitDir = "/mnt/uTorrent/"
		flags.Config = usr.HomeDir + sep + ".config" + sep + "qBittorrent" + sep + "qBittorrent.conf"
		flags.QBitDir = usr.HomeDir + sep + ".local" + sep + "share" + sep + "data" + sep + "qBittorrent" + sep + "BT_backup" + sep
	case "darwin":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		flags.BitDir = usr.HomeDir + sep + "Library" + sep + "Application Support" + sep + "uTorrent" + sep
		flags.Config = usr.HomeDir + sep + ".config" + sep + "qBittorrent" + sep + "qbittorrent.ini"
		flags.QBitDir = usr.HomeDir + sep + "Library" + sep + "Application Support" + sep + "QBittorrent" + sep + "BT_backup" + sep
	}

	_, err := goflags.Parse(&flags)
	if err != nil {
		os.Exit(0) // https://godoc.org/github.com/jessevdk/go-flags#ErrorType
	}

	if len(flags.Replace) != 0 {
		for _, str := range flags.Replace {
			patterns := strings.Split(str, ",")
			if len(patterns) < 2 {
				log.Println("Bad replace pattern")
				time.Sleep(30 * time.Second)
				os.Exit(1)
			}
		}
	}

	if flags.BitDir[len(flags.BitDir)-1] != os.PathSeparator {
		flags.BitDir += string(os.PathSeparator)
	}
	if flags.QBitDir[len(flags.QBitDir)-1] != os.PathSeparator {
		flags.QBitDir += string(os.PathSeparator)
	}

	if _, err := os.Stat(flags.BitDir); os.IsNotExist(err) {
		log.Println("Can't find uTorrent\\Bittorrent folder")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	if _, err := os.Stat(flags.QBitDir); os.IsNotExist(err) {
		log.Println("Can't find qBittorrent folder")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	resumefilepath := flags.BitDir + "resume.dat"
	if _, err := os.Stat(resumefilepath); os.IsNotExist(err) {
		log.Println("Can't find uTorrent\\Bittorrent resume file")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	resumefile, err := decodetorrentfile(resumefilepath)
	if err != nil {
		log.Println("Can't decode uTorrent\\Bittorrent resume file")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	if flags.WithoutTags == false {
		if _, err := os.Stat(flags.Config); os.IsNotExist(err) {
			fmt.Println("Can not read qBittorrent config file. Try run and close qBittorrent if you have not done" +
				" so already, or specify the path explicitly or do not import tags")
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
	}
	totaljobs := len(resumefile)
	chans := Channels{comChannel: make(chan string, totaljobs),
		errChannel:     make(chan string, totaljobs),
		boundedChannel: make(chan bool, runtime.GOMAXPROCS(0)*2)}
	color.Green("It will be performed processing from directory %v to directory %v\n", flags.BitDir, flags.QBitDir)
	color.HiRed("Check that the qBittorrent is turned off and the directory %v and config %v is backed up.\n\n",
		flags.QBitDir, flags.Config)
	fmt.Println("Press Enter to start")
	fmt.Scanln()
	log.Println("Started")
	transfertorrents(chans, flags, resumefile, totaljobs)
	fmt.Println("\nPress Enter to exit")
	fmt.Scanln()

}

func transfertorrents(chans Channels, flags Flags, resumefile map[string]interface{}, totaljobs int) {
	numjob := 1
	var oldtags string
	var newtags []string
	var wg sync.WaitGroup

	positionnum := 0
	for key, value := range resumefile {
		if key != ".fileguard" && key != "rec" {
			positionnum++
			if flags.WithoutTags == false {
				if labels, ok := value.(map[string]interface{})["labels"]; ok {
					for _, label := range labels.([]interface{}) {
						if len(label.(string)) > 0 {
							if ok, tag := checknotexists(ASCIIconvert(label.(string)), newtags); ok {
								newtags = append(newtags, tag)
							}
						}
					}
				}
			}
			wg.Add(1)
			chans.boundedChannel <- true
			go logic(key, value.(map[string]interface{}), &flags, &chans, positionnum, &wg)
		} else {
			totaljobs--
		}
	}
	go func() {
		wg.Wait()
		close(chans.comChannel)
		close(chans.errChannel)
	}()
	for message := range chans.comChannel {
		fmt.Printf("%v/%v %v \n", numjob, totaljobs, message)
		numjob++
	}
	var waserrors bool
	for message := range chans.errChannel {
		fmt.Printf("%v/%v %v \n", numjob, totaljobs, message)
		waserrors = true
		numjob++
	}
	if flags.WithoutTags == false {
		cfg, err := ini.Load(flags.Config)
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
				if ok, t := checknotexists(tag, newtags); ok {
					newtags = append(newtags, t)
				}
			}
			cfg.Section("BitTorrent").Key("Session\\Tags").SetValue(strings.Join(newtags, ", "))
		} else {
			cfg.Section("BitTorrent").NewKey("Session\\Tags", strings.Join(newtags, ", "))
		}
		cfg.SaveTo(flags.Config)
	}
	fmt.Println()
	log.Println("Ended")
	if waserrors {
		log.Println("Not all torrents was processed")
	}
}
