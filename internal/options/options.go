package options

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Opts struct {
	BitDir        string   `short:"s" long:"source" description:"Source directory that contains resume.dat and torrents files"`
	QBitDir       string   `short:"d" long:"destination" description:"Destination directory BT_backup (as default)"`
	Config        string   `short:"c" long:"config" description:"qBittorrent config file (for write tags)"`
	WithoutLabels bool     `long:"without-labels" description:"Do not export/import labels"`
	WithoutTags   bool     `long:"without-tags" description:"Do not export/import tags"`
	SearchPaths   []string `short:"t" long:"search" description:"Additional search path for torrents files\n	Example: --search='/mnt/olddisk/savedtorrents' --search='/mnt/olddisk/workstorrents'"`
	Replaces      []string `short:"r" long:"replace" description:"Replace paths.\n	Delimiter for from/to is comma - ,\n	Example: -r \"D:\\films,/home/user/films\" -r \"D:\\music,/home/user/music\"\n"`
	PathSeparator string   `long:"sep" description:"Default path separator that will use in all paths. You may need use this flag if you migrating from windows to linux in some cases"`
}

func PrepareOpts() *Opts {
	opts := &Opts{PathSeparator: string(os.PathSeparator)}
	switch OS := runtime.GOOS; OS {
	case "windows":
		opts.BitDir = filepath.Join(os.Getenv("APPDATA"), "uTorrent")
		opts.Config = filepath.Join(os.Getenv("APPDATA"), "qBittorrent", "qBittorrent.ini")
		opts.QBitDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "qBittorrent", "BT_backup")
	case "linux":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		opts.BitDir = "/mnt/uTorrent/"
		opts.Config = filepath.Join(usr.HomeDir, ".config", "qBittorrent", "qBittorrent.conf")
		opts.QBitDir = filepath.Join(usr.HomeDir, ".local", "share", "data", "qBittorrent", "BT_backup")
	case "darwin":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		opts.BitDir = filepath.Join(usr.HomeDir, "Library", "Application Support", "uTorrent")
		opts.Config = filepath.Join(usr.HomeDir, ".config", "qBittorrent", "qbittorrent.ini")
		opts.QBitDir = filepath.Join(usr.HomeDir, "Library", "Application Support", "QBittorrent", "BT_backup")
	}
	return opts
}

func ParseOpts(opts *Opts) *Opts {
	if _, err := flags.Parse(opts); err != nil { // https://godoc.org/github.com/jessevdk/go-flags#ErrorType
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Println(err)
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
	}
	return opts
}

func HandleOpts(opts *Opts) {
	opts.SearchPaths = append(opts.SearchPaths, opts.BitDir)
}

func OptsCheck(opts *Opts) error {
	if len(opts.Replaces) != 0 {
		for _, str := range opts.Replaces {
			patterns := strings.Split(str, ",")
			if len(patterns) != 2 {
				return fmt.Errorf("Bad replace pattern")
			}
		}
	}

	if _, err := os.Stat(opts.BitDir); os.IsNotExist(err) {
		return fmt.Errorf("Can't find uTorrent\\Bittorrent folder")
	}

	if _, err := os.Stat(opts.QBitDir); os.IsNotExist(err) {
		return fmt.Errorf("Can't find qBittorrent folder")
	}

	if opts.WithoutTags == false {
		if _, err := os.Stat(opts.Config); os.IsNotExist(err) {
			return fmt.Errorf("Can not read qBittorrent config file. Try run and close qBittorrent if you have not done" +
				" so already, or specify the path explicitly or do not import tags")
		}
	}
	if runtime.GOOS == "linux" {
		if opts.SearchPaths == nil {
			return fmt.Errorf("On linux systems you must define search path for torrents")
		}
	}
	return nil
}

func MakeOpts() *Opts {
	opts := PrepareOpts()
	ParseOpts(opts)
	HandleOpts(opts)
	err := OptsCheck(opts)
	if err != nil {
		log.Println(err)
		time.Sleep(time.Duration(30) * time.Second)
		os.Exit(1)
	}
	return opts
}
