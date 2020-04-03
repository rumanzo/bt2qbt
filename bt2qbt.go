package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-ini/ini"
	"github.com/juju/gnuflag"
	"github.com/rumanzo/bt2qbt/libtorrent"
	"github.com/rumanzo/bt2qbt/replace"
	"github.com/zeebo/bencode"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
	bitDir, qBitDir, config    string
	withoutLabels, withoutTags bool
	replace                    string
}

type Channels struct {
	comChannel     chan string
	errChannel     chan string
	boundedChannel chan bool
}

func (channels *Channels) MakeApiHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		channels.comChannel <- r.Host
		io.WriteString(w, "Hello world!")
	}
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
		UploadRateLimit: 0, QbtName: "", WithoutLabels: flags.withoutLabels, WithoutTags: flags.withoutTags}
	if ok := filepath.IsAbs(key); ok {
		newstructure.TorrentFilePath = key
	} else {
		newstructure.TorrentFilePath = flags.bitDir + key
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

	if flags.replace != "" {
		for _, str := range strings.Split(flags.replace, ";") {
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

	if err = encodetorrentfile(flags.qBitDir+newbasename+".fastresume", &newstructure); err != nil {
		chans.errChannel <- fmt.Sprintf("Can't create qBittorrent fastresume file %v", flags.qBitDir+newbasename+".fastresume")
		return err
	}
	if err = copyfile(newstructure.TorrentFilePath, flags.qBitDir+newbasename+".torrent"); err != nil {
		chans.errChannel <- fmt.Sprintf("Can't create qBittorrent torrent file %v", flags.qBitDir+newbasename+".torrent")
		return err
	}
	chans.comChannel <- fmt.Sprintf("Sucessfully imported %v", key)
	return nil
}

func IndexRouterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %", "s")
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
		flags.bitDir = os.Getenv("APPDATA") + sep + "uTorrent" + sep
		flags.config = os.Getenv("APPDATA") + sep + "qBittorrent" + sep + "qBittorrent.ini"
		flags.qBitDir = os.Getenv("LOCALAPPDATA") + sep + "qBittorrent" + sep + "BT_backup" + sep
	case "linux":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		flags.bitDir = "/mnt/uTorrent/"
		flags.config = usr.HomeDir + sep + ".config" + sep + "qBittorrent" + sep + "qBittorrent.conf"
		flags.qBitDir = usr.HomeDir + sep + ".local" + sep + "share" + sep + "data" + sep + "qBittorrent" + sep + "BT_backup" + sep
	case "darwin":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		flags.bitDir = usr.HomeDir + sep + "Library" + sep + "Application Support" + sep + "uTorrent" + sep
		flags.config = usr.HomeDir + sep + ".config" + sep + "qBittorrent" + sep + "qbittorrent.ini"
		flags.qBitDir = usr.HomeDir + sep + "Library" + sep + "Application Support" + sep + "QBittorrent" + sep + "BT_backup" + sep
	}

	gnuflag.StringVar(&flags.bitDir, "source", flags.bitDir,
		"Source directory that contains resume.dat and torrents files")
	gnuflag.StringVar(&flags.bitDir, "s", flags.bitDir,
		"Source directory that contains resume.dat and torrents files")
	gnuflag.StringVar(&flags.qBitDir, "destination", flags.qBitDir,
		"Destination directory BT_backup (as default)")
	gnuflag.StringVar(&flags.qBitDir, "d", flags.qBitDir,
		"Destination directory BT_backup (as default)")
	gnuflag.StringVar(&flags.config, "qconfig", flags.config,
		"qBittorrent config files (for write tags)")
	gnuflag.StringVar(&flags.config, "c", flags.config,
		"qBittorrent config files (for write tags)")
	gnuflag.BoolVar(&flags.withoutLabels, "without-labels", false, "Do not export/import labels")
	gnuflag.BoolVar(&flags.withoutTags, "without-tags", false, "Do not export/import tags")
	gnuflag.StringVar(&flags.replace, "replace", "", "Replace paths.\n	"+
		"Delimiter for replaces - ;\n	"+
		"Delimiter for from/to - ,\n	Example: \"D:\\films,/home/user/films;\\,/\"\n	"+
		"If you use path separator different from you system, declare it mannually")
	gnuflag.Parse(true)

	if flags.replace != "" {
		for _, str := range strings.Split(flags.replace, ";") {
			patterns := strings.Split(str, ",")
			if len(patterns) < 2 {
				log.Println("Bad replace pattern")
				time.Sleep(30 * time.Second)
				os.Exit(1)
			}
		}
	}

	if flags.bitDir[len(flags.bitDir)-1] != os.PathSeparator {
		flags.bitDir += string(os.PathSeparator)
	}
	if flags.qBitDir[len(flags.qBitDir)-1] != os.PathSeparator {
		flags.qBitDir += string(os.PathSeparator)
	}

	if _, err := os.Stat(flags.bitDir); os.IsNotExist(err) {
		log.Println("Can't find uTorrent\\Bittorrent folder")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	if _, err := os.Stat(flags.qBitDir); os.IsNotExist(err) {
		log.Println("Can't find qBittorrent folder")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}
	resumefilepath := flags.bitDir + "resume.dat"
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
	if flags.withoutTags == false {
		if _, err := os.Stat(flags.config); os.IsNotExist(err) {
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
	http.HandleFunc("/", chans.MakeApiHandler())
	go http.ListenAndServe(":9000", nil)
	openBrowser("http://localhost:9000")
	fmt.Println(<-chans.comChannel)
	color.Green("It will be performed processing from directory %v to directory %v\n", flags.bitDir, flags.qBitDir)
	color.HiRed("Check that the qBittorrent is turned off and the directory %v and config %v is backed up.\n\n",
		flags.qBitDir, flags.config)
	fmt.Println("Press Enter to start")
	fmt.Scanln()
	log.Println("Started")
	transfertorrents(chans, flags, resumefile, totaljobs)

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
			if flags.withoutTags == false {
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
	if flags.withoutTags == false {
		cfg, err := ini.Load(flags.config)
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
		cfg.SaveTo(flags.config)
	}
	fmt.Println()
	log.Println("Ended")
	if waserrors {
		log.Println("Not all torrents was processed")
	}
	fmt.Println("\nPress Enter to exit")
	fmt.Scanln()
}
