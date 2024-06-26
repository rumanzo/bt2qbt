package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/internal/transfer"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"github.com/zeebo/bencode"
)

var version, commit, date, buildImage string

func main() {
	opts := options.MakeOpts()

	if opts.Version {
		fmt.Printf("Version: %v\nCommit: %v\nGolang version: %v\nBuild image: %v\n", version, commit, runtime.Version(), buildImage)
		os.Exit(0)
	}

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
	// hate utorrent for heterogeneous resume.dat scheme
	delete(resumeFile, ".fileguard")
	delete(resumeFile, "rec")
	b, _ := bencode.EncodeBytes(resumeFile)
	resumeItems := map[string]*utorrentStructs.ResumeItem{}
	err = bencode.DecodeBytes(b, &resumeItems)
	if err != nil {
		log.Printf("Can't convert resume.dat. Err: %v\n", err)
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}

	color.Green("It will be performed processing from directory %v to directory %v\n", opts.BitDir, opts.QBitDir)
	color.HiRed("Check that the qBittorrent is turned off and the directory %v and %v is backed up.\n",
		opts.QBitDir, opts.Categories)
	color.HiRed("Check that you previously disable option \"Append .!ut/.!bt to incomplete files\" in preferences of uTorrent/Bittorrent \n")
	color.HiRed("Close uTorrent/Bittorrent and qBittorrent previously\n\n")
	fmt.Println("Press Enter to start")
	fmt.Scanln()
	log.Println("Started")

	transfer.HandleResumeItems(opts, resumeItems)

	fmt.Println("\nPress Enter to exit")
	fmt.Scanln()

}
