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
		fmt.Println("We've failed to open the data store.")
		log.WithFields(log.Fields{
			"err": err,
		}).Error("We've failed to open the data store.")
		return err
	}
	defer store.Close()

	// Fetch the seasons/episodes associated with the found show.
	persistedShow := store.NewShow(show.Source, show.ID, show.URL, show.Title)
	err = ui.Lookup(persistedShow)
	if err != nil {
		fmt.Println("We've encountered a problem looking up seasons for the show.")
		log.WithFields(log.Fields{
			"err": err,
		}).Error("We've encountered a problem looking up seasons for the show.")
		return err
	}

	if len(persistedShow.Episodes()) == 0 {
		fmt.Println("No episodes could be found for show.")
		log.WithFields(log.Fields{
			"show": persistedShow.Title,
		}).Info("No episodes could be found for show.")
		return nil
	}

	err = store.CreateShow(persistedShow)
	if err != nil {
		fmt.Println("Show already exists. Remove it or search for something else. If you want to update it do: getme -u")
		log.WithFields(log.Fields{
			"err":  err,
			"show": persistedShow.Title,
		}).Fatal("Show already exists.")
	}

	if !noDownload {
		downloadTorrents(persistedShow)
	}

	return nil
}

func downloadTorrents(show *store.Show) {
	torrents, err := ui.SearchTorrents(show)
	if err != nil {
		// But that doesn't mean nothing worked...
		fmt.Println("Something went wrong looking for your torrents. Continuing nonetheless")
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("Something went wrong looking for your torrents. Continuing nonetheless")
	}
	if len(torrents) == 0 {
		fmt.Println("Didn't find any torrents for show.")
		log.WithFields(log.Fields{
			"show": show.Title,
		}).Info("Didn't find any torrents for show.")
	}
	err = ui.Download(torrents)
	if err != nil {
		fmt.Println("Something went wrong downloading a torrent.")
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("Something went wrong downloading a torrent.")
	}
	ui.DisplayPendingEpisodes(show)
}

func loadConfig() {
	ui.EnsureConfig()
}

var update bool
var mediaName string
var logLevel int
var noDownload bool
var version bool
var versionNumber = "0.2"

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s <flags>\n", os.Args[0])
		flag.PrintDefaults()
	}
	const (
		addUsage        = "The name of the show/movie to add."
		updateUsage     = "Update the already added shows/movies and download pending torrents."
		logLevelUsage   = "Set log level (0,1,2,3,4,5, higher is more logging)."
		noDownloadUsage = "Find the show but don't download the torrents."
		versionUsage    = "Show version"
	)

	flag.StringVar(&mediaName, "add", "", addUsage)
	flag.StringVar(&mediaName, "a", "", addUsage+" (shorthand)")

	flag.BoolVar(&update, "update", false, updateUsage)
	flag.BoolVar(&update, "u", false, updateUsage+" (shorthand)")

	flag.IntVar(&logLevel, "log-level", int(log.ErrorLevel), logLevelUsage)
	flag.IntVar(&logLevel, "l", int(log.ErrorLevel), logLevelUsage+" (shorthand)")

	flag.BoolVar(&noDownload, "no-download", false, noDownloadUsage)
	flag.BoolVar(&noDownload, "n", false, noDownloadUsage+" (shorthand)")

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
		fmt.Println("We've failed to open the data store.")
		log.WithFields(log.Fields{
			"err": err,
		}).Error("We've failed to open the data store.")
		return
	}
	defer store.Close()

	ui.Update(store)
}

func allEmpty(results []sources.SearchResult) bool {
	for _, result := range results {
		if len(result.Shows) > 0 {
			return false
		}
	}
	return true
}

// TODO shouldn't this be in the ui package?
func addMedia() {
	if mediaName == "" {
		fmt.Println("Please specify a name to add. Like so: ./getme -a 'My show'.")
		return
	}

	matches := ui.Search(mediaName)
	if allEmpty(matches) {
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
		log.Panic("Match is neither a Show or a Movie")
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

	config.SetLoggerTo(logLevel)

	if update {
		updateMedia()
	} else {
		addMedia()
	}
}
