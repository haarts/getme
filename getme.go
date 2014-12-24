package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
	"github.com/haarts/getme/ui"
)

func handleShow(show *sources.Show) error {
	store := store.Open(config.StateDir)
	defer store.Close()

	// Fetch the seasons/episodes associated with the found show.
	err := ui.Lookup(show)
	if err != nil {
		fmt.Println("We've encountered a problem looking up seasons for the show. The error:")
		fmt.Println(" ", err)
		return err
	}

	if len(show.Episodes()) == 0 {
		fmt.Printf("No episodes could be found for %s.", show.DisplayTitle())
		return nil
	}

	store.CreateShow(show)

	// We have two entry points. One on the first run and one when running as daemon.
	// So we create episodes based on seasons always. Then look at the disk/store and figure out
	// what is missing.

	torrents, err := ui.SearchTorrents(show.Episodes())
	if err != nil {
		fmt.Println("Something went wrong looking for your torrents:", err)
		return err
	}
	if len(torrents) == 0 {
		fmt.Println("Didn't find any torrents for", show.DisplayTitle())
		return nil
	}
	err = ui.Download(torrents, config.WatchDir)
	if err != nil {
		fmt.Println("Something went wrong downloading a torrent:", err)
	}
	ui.DisplayPendingEpisodes(show)

	return nil
}

func configFilePath() string {
	var u *user.User
	if u, _ = user.Current(); u == nil {
		return ""
	}
	dirPath := path.Join(u.HomeDir, ".config", "getme") // TODO What's the sane location for Windows?
	filePath := path.Join(dirPath, "config.ini")
	return filePath
}

func checkConfig() error {
	_, err := os.Stat(configFilePath())
	return err
}

func writeDefaultConfig() {
	f := configFilePath()
	dir, _ := path.Split(f)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	if err := ioutil.WriteFile(f, defaultConfigData(dir), 0644); err != nil {
		return
	}
}

func defaultConfigData(homeDir string) []byte {
	watchDir := fmt.Sprintln("watch_dir = /tmp/torrents")
	return []byte(watchDir + fmt.Sprintf("state_dir = %sstate\n", homeDir))
}

var config Config

type Config struct {
	WatchDir string
	StateDir string
}

func readConfig() (Config, error) {
	file, err := os.Open(configFilePath())
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	conf := Config{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		parts := strings.Split(text, "=")
		for i, _ := range parts {
			parts[i] = strings.Trim(parts[i], " ")
		}
		switch parts[0] {
		case "watch_dir":
			conf.WatchDir = parts[1]
		case "state_dir":
			conf.StateDir = parts[1]
		default:
			return Config{}, errors.New("Found an unknown key in config.ini: " + parts[0])
		}
	}

	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	return conf, nil
}

func main() {
	err := checkConfig()
	if err != nil && os.IsNotExist(err) {
		fmt.Println("It seems that there is no config file present at", configFilePath())
		fmt.Println("Writing a default one, please inspect it and restart GetMe.")
		writeDefaultConfig()
		return
	}
	conf, err := readConfig()
	if err != nil {
		fmt.Println("Something went wrote reading the config file:", err)
	}
	config = conf

	matches, errors := ui.Search(ui.GetQuery())
	if errors != nil {
		fmt.Println("We've encountered a problem searching. The error:")
		fmt.Println(" ", errors)
	}
	if len(matches) == 0 {
		fmt.Println("We haven't found what you were looking for.")
		return
	}

	// Determine which show/movie ppl want.
	match := ui.DisplayBestMatchConfirmation(matches)
	if match == nil {
		match = ui.DisplayAlternatives(matches)
	}

	if match == nil {
		fmt.Println("We're sorry we couldn't find it for you.")
		return
	}

	switch m := (*match).(type) {
	case sources.Show:
		err := handleShow(&m)
		if err != nil {
			return
		}
	case sources.Movie:
	// TODO Handle 'Movie' case.

	default:
		panic("Match is neither a Show or a Movie")
	}

	fmt.Println("All done!")
	return
}
