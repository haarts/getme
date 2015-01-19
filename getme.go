// GetMe makes it easy to download and monitor movies and TV shows.
package main

import (
	"flag"
	"fmt"

	"github.com/haarts/getme/config"
	"github.com/haarts/getme/store"
	"github.com/haarts/getme/ui"
)

var log = config.Log()

func handleShow(show *store.Show) error {
	store, err := store.Open()
	if err != nil {
		fmt.Println("We've failed to open the data store. The error:")
		fmt.Println(" ", err)
		return err
	}
	defer store.Close()

	// Fetch the seasons/episodes associated with the found show.
	err = ui.Lookup(show)
	if err != nil {
		fmt.Println("We've encountered a problem looking up seasons for the show. The error:")
		fmt.Println(" ", err)
		return err
	}

	if len(show.Episodes()) == 0 {
		fmt.Printf("No episodes could be found for %s.", show.DisplayTitle())
		return nil
	}

	err = store.CreateShow(show)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Remove it or search for something else. If you want to update it do: getme -u")
		log.Fatalf("Show '%s' already exists", show.Title)
	}

	// We have two entry points. One on the first run and one when running as daemon.
	// So we create episodes based on seasons always. Then look at the disk/store and figure out
	// what is missing.

	torrents, err := ui.SearchTorrents(show)
	if err != nil {
		fmt.Println("Something went wrong looking for your torrents:", err) // But that doesn't mean nothing worked...
		fmt.Println("Continuing nonetheless.")
	}
	if len(torrents) == 0 {
		fmt.Println("Didn't find any torrents for", show.DisplayTitle())
		return nil
	}
	err = ui.Download(torrents)
	if err != nil {
		fmt.Println("Something went wrong downloading a torrent:", err)
	}
	ui.DisplayPendingEpisodes(show)

	return nil
}

func loadConfig() {
	ui.EnsureConfig()
}

var update bool
var mediaName string
var debug bool

func init() {
	const (
		addUsage    = "The name of the show/movie to add."
		updateUsage = "Update the already added shows/movies and download pending torrents."
		debugUsage  = "Turn on debugging output"
	)

	flag.StringVar(&mediaName, "add", "", addUsage)
	flag.StringVar(&mediaName, "a", "", addUsage+" (shorthand)")

	flag.BoolVar(&update, "update", false, updateUsage)
	flag.BoolVar(&update, "u", false, updateUsage+" (shorthand)")

	flag.BoolVar(&debug, "debug", false, debugUsage)
	flag.BoolVar(&debug, "D", false, debugUsage+" (shorthand)")

	// TODO add a remove flag. (could just remove the file in stateDir)
	//flag.BoolVar(&remove, "remove", false, removeUsage))
	//flag.BoolVar(&remove, "r", false, removeUsage+" (shorthand)")

	// TODO add a quiet flag (-q)
	// TODO add a yes flag (-y)
}

func updateMedia() {
	store, err := store.Open()
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

	matches, errors := ui.Search(mediaName)
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
	case store.Show:
		err := handleShow(&m)
		if err != nil {
			return
		}
	case store.Movie:
	// TODO Handle 'Movie' case.

	default:
		panic("Match is neither a Show or a Movie")
	}

	fmt.Println("All done!")
	return
}

func main() {
	flag.Parse()

	loadConfig()
	if debug {
		config.SetLoggerToDebug()
	}

	if update {
		updateMedia()
	} else {
		addMedia()
	}
}
