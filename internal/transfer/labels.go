package transfer

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"os"
	"strings"
	"time"
)

func ProcessLabels(opts *options.Opts, newtags []string) {
	var oldtags string
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
	err = cfg.SaveTo(opts.Config)
	if err != nil {
		fmt.Printf("Unexpected error while save qBittorrent config.ini. Error:\n%v\n", err)
	}
}
