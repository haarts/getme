// GetMe makes it easy to download and monitor movies and TV shows.
package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/config"
	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
	"github.com/haarts/getme/ui"
)

func handleShow(show *sources.Show) error {
	store, err := store.Open(config.Config().StateDir)
	if err != nil {
		fmt.Println("We've failed to open the data store. The error:")
		fmt.Println(" ", err)
		return err
	}
	defer store.Close()

	// Fetch the seasons/episodes associated with the found show.
	persistedShow := store.NewShow(show.Source, show.ID, show.Title)
	err = ui.Lookup(persistedShow)
	if err != nil {
		fmt.Println("We've encountered a problem looking up seasons for the show. The error:")
		fmt.Println(" ", err)
		return err
	}

	if len(persistedShow.Episodes()) == 0 {
		fmt.Printf("No episodes could be found for %s.\n", persistedShow.Title)
		return nil
	}

	err = store.CreateShow(persistedShow)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Remove it or search for something else. If you want to update it do: getme -u")
		log.Fatalf("Show '%s' already exists", persistedShow.Title)
	}

	// We have two entry points. One on the first run and one when running as daemon.
	// So we create episodes based on seasons always. Then look at the disk/store and figure out
	// what is missing.

	torrents, err := ui.SearchTorrents(persistedShow)
	if err != nil {
		fmt.Println("Something went wrong looking for your torrents:", err) // But that doesn't mean nothing worked...
		fmt.Println("Continuing nonetheless.")
	}
	if len(torrents) == 0 {
		fmt.Println("Didn't find any torrents for", persistedShow.Title)
		return nil
	}
	err = ui.Download(torrents)
	if err != nil {
		fmt.Println("Something went wrong downloading a torrent:", err)
	}
	ui.DisplayPendingEpisodes(persistedShow)

	return nil
}

func loadConfig() {
	ui.EnsureConfig()
}

var update bool
var mediaName string
var debug bool
var version bool
var versionNumber = "0.2"

func init() {
	const (
		addUsage     = "The name of the show/movie to add."
		updateUsage  = "Update the already added shows/movies and download pending torrents."
		debugUsage   = "Turn on debugging output"
		versionUsage = "Show version"
	)

	flag.StringVar(&mediaName, "add", "", addUsage)
	flag.StringVar(&mediaName, "a", "", addUsage+" (shorthand)")

	flag.BoolVar(&update, "update", false, updateUsage)
	flag.BoolVar(&update, "u", false, updateUsage+" (shorthand)")

	flag.BoolVar(&debug, "debug", false, debugUsage)
	flag.BoolVar(&debug, "D", false, debugUsage+" (shorthand)")

	flag.BoolVar(&version, "version", false, versionUsage)
	flag.BoolVar(&version, "v", false, versionUsage+" (shorthand)")

	// TODO add a remove flag. (could just remove the file in stateDir)
	//flag.BoolVar(&remove, "remove", false, removeUsage))
	//flag.BoolVar(&remove, "r", false, removeUsage+" (shorthand)")

	// TODO add a quiet flag (-q)
	// TODO add a yes flag (-y)
}

func updateMedia() {
	store, err := store.Open(config.Config().StateDir)
	if err != nil {
		fmt.Println("We've failed to open the data store. The error:")
		fmt.Println(" ", err)
		return
	}
	defer store.Close()

	ui.Update(store)
}

func addMedia() {
	if mediaName == "" {
		fmt.Println("Please specify a name to add. Like so: ./getme -a 'My show'.")
		return
	}

	matches := ui.Search(mediaName)
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

	switch m := (match).(type) {
	case *sources.Show:
		err := handleShow(m)
		if err != nil {
			return
		}
	case *store.Movie:
	// TODO Handle 'Movie' case.

	default:
		panic("Match is neither a Show or a Movie")
	}

	fmt.Println("All done!")
	return
}

func main() {
	flag.Parse()

	if version {
		fmt.Printf("Version %s\n", versionNumber)
		os.Exit(1)
	}

	loadConfig()
	config.SetLoggerOutput(config.Config().LogDir)

	if debug {
		config.SetLoggerToDebug()
	}

	if update {
		updateMedia()
	} else {
		addMedia()
	}
}
