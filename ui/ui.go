// Package ui handles the interaction with the user on the CLI.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/config"
	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
	"github.com/haarts/getme/torrents"
)

// NOTE no log calls should appear here. That stuff should be handled in the
// underlying layer.

var conf = config.Config()

// EnsureConfig tries to load the config file. If there is no such file it will
// create one and exits.
func EnsureConfig() {
	err := config.CheckConfig()
	if err != nil && os.IsNotExist(err) {
		fmt.Println("It seems that there is no config file present at ", config.ConfigFile())
		fmt.Println("Writing a default one, please inspect it and restart GetMe.")
		config.WriteDefaultConfig()
		os.Exit(1)
	}

	if config.Config() == nil {
		os.Exit(1)
	}
}

// DisplayPendingEpisodes shows, on stdout, the episodes pending for a
// particular show.
func DisplayPendingEpisodes(show *store.Show) {
	xs := show.PendingSeasons()
	for _, x := range xs {
		fmt.Printf("Pending: %s season %d\n", show.Title, x.Season)
	}
	ys := show.PendingEpisodes()
	if len(ys) > 10 {
		fmt.Println("<snip>")
		ys = ys[len(ys)-10:]
	}
	for _, y := range ys {
		fmt.Printf("Pending: %s season %d episode %d\n", show.Title, y.Season(), y.Episode)
	}
}

// DisplayBestMatchConfirmation asks the user to confirm the, what we THINK, is
// the best match.
func DisplayBestMatchConfirmation(matches []sources.SearchResult) sources.Match {
	nonNilMatch := firstNonNilMatch(matches)
	if nonNilMatch == nil {
		return nil
	}

	displayBestMatch(nonNilMatch)
	fmt.Print("Is this the one you want? [Y/n] ")
	line := getUserInput()

	if line == "" || line == "y" || line == "Y" {
		return nonNilMatch
	}
	return nil
}

func firstNonNilMatch(matches []sources.SearchResult) sources.Match {
	for _, ms := range matches {
		if len(ms.Shows) != 0 {
			return &ms.Shows[0]
		}
	}
	return nil // This really shouldn't happen.
}

// DisplayAlternatives shows as many lists as there are sources with found
// matches. The user is asked to select one of them.
// TODO break this func up. Too long.
func DisplayAlternatives(ms []sources.SearchResult) sources.Match {
	fmt.Println("Which one ARE you looking for?")
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	var names []string
	for _, m := range ms {
		names = append(names, m.Name)
	}
	fmt.Fprint(w, strings.Join(names, "\t")+"\n")

	var generators []func() (string, []interface{})
	step := 1
	for i, m := range ms {
		if i > 0 {
			step += len(ms[i-1].Shows)
		}
		generators = append(generators, createGenerator(m.Shows, step))
	}

	anyGeneratorAlive := true
	for anyGeneratorAlive {
		var collectorFmt []string
		var collectorArgs []interface{}
		anyGeneratorAlive = false
		for _, g := range generators {
			fmtString, args := g()
			if fmtString != "" {
				anyGeneratorAlive = true
			}
			collectorFmt = append(collectorFmt, fmtString)
			collectorArgs = append(collectorArgs, args...)
		}
		fmt.Fprintf(w, strings.Join(collectorFmt, "\t")+"\n", collectorArgs...)
	}

	w.Flush()

	fmt.Print("Enter the correct number: ")
	line := getUserInput()

	// User abort
	if line == "" {
		return nil
	}

	i, err := strconv.Atoi(line)
	// User mis-typed, try again
	if err != nil {
		fmt.Printf("Didn't understand '%s'. Try again (ENTER quits).\n", line)
		return DisplayAlternatives(ms)
	}

	var flatList []sources.Show
	for _, m := range ms {
		flatList = append(flatList, m.Shows...)
	}

	return &flatList[i-1]
}

func createGenerator(ms []sources.Show, step int) func() (string, []interface{}) {
	i := 0
	f := func() (string, []interface{}) {
		var fmtString string
		var args []interface{}
		if i < len(ms) {
			fmtString = "[%d] %s"
			args = []interface{}{i + step, ms[i].DisplayTitle()}
		}
		i++
		return fmtString, args
	}

	return f
}

// Download goes about downloading torrents found based on the pending
// episodes/seasons.
func Download(foundTorrents []torrents.Torrent) error {
	fmt.Printf("Downloading %d torrents", len(foundTorrents))
	c := startProgressBar()
	defer stopProgressBar(c)

	return torrents.Download(foundTorrents, conf.WatchDir)
}

// SearchTorrents provides some feedback to the user and searches for torrents
// for the pending items.
func SearchTorrents(show *store.Show) ([]torrents.Torrent, error) {
	fmt.Printf(
		"Searching for %d torrents",
		len(show.PendingSeasons())+len(show.PendingEpisodes()),
	)

	c := startProgressBar()
	defer stopProgressBar(c)

	return torrents.Search(show)
}

// Search converts a user provided search string into a linked list of
// potential matches.
func Search(query string) []sources.SearchResult {
	fmt.Printf("Seaching for '%s' on: ", query)
	fmt.Print(strings.Join(sources.SourceNames(), ", "))
	fmt.Print("\n")

	c := startProgressBar()
	defer stopProgressBar(c)

	matches := sources.Search(query)
	if isAnyErrorless(matches) { // Silently ignore all errors as long as 1 succeeded.
		return matches
	}

	return nil
}

// Lookup takes a show previously selected by the user and finds the seasons
// and episodes with it.
func Lookup(show *store.Show) error {
	fmt.Printf("Looking up seasons and episodes for '%s'", show.Title)
	c := startProgressBar()
	defer stopProgressBar(c)

	return sources.UpdateSeasonsAndEpisodes(show)
}

// Update takes all the shows stored on disk and adds any new episodes to them.
func Update(store *store.Store) {
	fmt.Println("Updating media from sources and downloading pending torrents.")

	updateShows(store.Shows())
	updateMovies(store.Movies())
}

func updateShows(shows map[string]*store.Show) {
	for _, show := range shows {
		err := updateShow(show)
		if err != nil {
			fmt.Printf("Error updating '%s': %s\n", show.Title, err.Error())
			continue
		}

		torrents, err := SearchTorrents(show)
		if err != nil {
			fmt.Printf("Error searching torrents for '%s': %s\n", show.Title, err.Error())
			continue
		}

		err = Download(torrents)
		if err != nil {
			fmt.Printf("Error downloading torrents for '%s': %s\n", show.Title, err.Error())
			continue
		}

		DisplayPendingEpisodes(show)
		fmt.Println("")
	}
}

func updateShow(show *store.Show) error {
	fmt.Printf("Updating '%s'", show.Title)

	c := startProgressBar()
	defer stopProgressBar(c)

	return sources.UpdateSeasonsAndEpisodes(show)
}

// TODO this is easier since we don't have to check for new episodes etc. Just pending.
func updateMovies(movies map[string]*store.Movie) {
	for _, movie := range movies {
		fmt.Printf("movie %+v\n", movie)
		// ... get updated info
		//store.UpdateMovie(updatedMovie)
	}
}

func isAnyErrorless(errors []sources.SearchResult) bool {
	for _, e := range errors {
		if e.Error == nil {
			return true
		}
	}
	return false
}

func isAllNil(errors []error) bool {
	for _, e := range errors {
		if e != nil {
			return false
		}
	}
	return true
}

func displayBestMatch(bestMatch sources.Match) {
	fmt.Println("The best match we found is:")
	fmt.Println(" ", bestMatch.DisplayTitle())
}

func startProgressBar() *time.Ticker {
	c := time.NewTicker(1 * time.Second)
	go func() {
		for _ = range c.C {
			fmt.Print(".")
		}
	}()

	return c
}

func stopProgressBar(c *time.Ticker) {
	c.Stop()
	fmt.Print("\n")
}

func getUserInput() string {
	bio := bufio.NewReader(os.Stdin)
	line, err := bio.ReadString('\n')
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error reading user input.")
	}
	return strings.Trim(line, "\r\n")
}
