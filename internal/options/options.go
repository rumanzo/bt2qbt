package options

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/rumanzo/bt2qbt/pkg/fileHelpers"
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
	Categories    string   `short:"c" long:"categories" description:"Path to qBittorrent categories.json file (for write tags)"`
	WithoutLabels bool     `long:"without-labels" description:"Do not export/import labels"`
	WithoutTags   bool     `long:"without-tags" description:"Do not export/import tags"`
	SearchPaths   []string `short:"t" long:"search" description:"Additional search path for torrents files\n	Example: --search='/mnt/olddisk/savedtorrents' --search='/mnt/olddisk/workstorrents'"`
	Replaces      []string `short:"r" long:"replace" description:"Replace save paths. Important: you have to use single slashes in paths\n	Delimiter for from/to is comma - ,\n	Example: -r \"D:/films,/home/user/films\" -r \"D:/music,/home/user/music\"\n"`
	PathSeparator string   `long:"sep" description:"Default path separator that will use in all paths. You may need use this flag if you migrating from windows to linux in some cases"`
	Version       bool     `short:"v" long:"version" description:"Show version"`
}

func PrepareOpts() *Opts {
	opts := &Opts{PathSeparator: string(os.PathSeparator)}
	switch OS := runtime.GOOS; OS {
	case "windows":
		opts.BitDir = filepath.Join(os.Getenv("APPDATA"), "uTorrent")
		opts.Categories = filepath.Join(os.Getenv("APPDATA"), "qBittorrent", "categories.json")
		opts.QBitDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "qBittorrent", "BT_backup")
	case "linux":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		opts.BitDir = "/mnt/uTorrent/"
		opts.Categories = filepath.Join(usr.HomeDir, ".config", "qBittorrent", "categories.json")
		opts.QBitDir = filepath.Join(usr.HomeDir, ".local", "share", "data", "qBittorrent", "BT_backup")
	case "darwin":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		opts.BitDir = filepath.Join(usr.HomeDir, "Library", "Application Support", "uTorrent")
		opts.Categories = filepath.Join(usr.HomeDir, ".config", "qBittorrent", "categories.json")
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

// HandleOpts used for enrichment opts after first creation
func HandleOpts(opts *Opts) {
	opts.SearchPaths = append(opts.SearchPaths, opts.BitDir)

	qbtDir := fileHelpers.Normalize(opts.QBitDir, `/`)
	if strings.Contains(qbtDir, `profile/qBittorrent/data/BT_backup`) {
		qbtRootDir, _ := strings.CutSuffix(qbtDir, `data/BT_backup`)

		// check that user not define categories
		refOpts := PrepareOpts()
		if refOpts.Categories == opts.Categories {
			opts.Categories = fileHelpers.Join([]string{qbtRootDir, `config/categories.json`}, opts.PathSeparator)
		}
	}
}

func OptsCheck(opts *Opts) error {
	if len(opts.Replaces) != 0 {
		for _, str := range opts.Replaces {
			patterns := strings.Split(str, ",")
			if len(patterns) != 2 {
				return fmt.Errorf("bad replace pattern")
			}
		}
	}

	if _, err := os.Stat(opts.BitDir); os.IsNotExist(err) {
		return fmt.Errorf("can't find uTorrent\\Bittorrent folder")
	}

	if _, err := os.Stat(opts.QBitDir); os.IsNotExist(err) {
		return fmt.Errorf("can't find qBittorrent folder")
	}

	if runtime.GOOS == "linux" {
		if opts.SearchPaths == nil {
			return fmt.Errorf("on linux systems you must define search path for torrents")
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
