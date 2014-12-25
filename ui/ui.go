package ui

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/haarts/getme/search_engines"
	"github.com/haarts/getme/sources"
)

func GetQuery() string {
	if len(os.Args) != 2 {
		fmt.Println("Please pass a search query.")
		os.Exit(1)
	}

	query := os.Args[1]
	return query
}

func DisplayPendingEpisodes(show *sources.Show) {
	es := show.PendingEpisodes()
	for _, e := range es {
		fmt.Println("Pending: ", e.QueryNames()[0])
	}
}

func DisplayBestMatchConfirmation(matches []sources.Match) *sources.Match {
	displayBestMatch(matches[0])
	fmt.Print("Is this the one you want? [Y/n] ")
	line := getUserInput()

	if line == "" || line == "y" || line == "Y" {
		return &matches[0]
	} else {
		return nil
	}
}

// TODO change arg type to [][]sources.Match
// Then make len([][]Match) generators: generator([i][]Match, len([i-][]Match))
// The second arg is to count the number for the user
func DisplayAlternatives(ms []sources.Match) *sources.Match {
	fmt.Println("Which one ARE you looking for?")
	for i, m := range ms {
		fmt.Printf("[%d] %s\n", i+1, m.DisplayTitle())
	}

	fmt.Print("Enter the correct number: ")
	line := getUserInput()

	// User abort
	if line == "" {
		return nil
	}

	i, err := strconv.Atoi(line)
	// User mis-typed, try again
	if err != nil {
		return DisplayAlternatives(ms)
	}

	return &ms[i-1]
}

func Download(torrents []search_engines.Torrent, watchDir string) (err error) {
	fmt.Printf("Downloading %d torrents", len(torrents))
	c := startProgressBar()
	defer stopProgressBar(c)

	for _, torrent := range torrents {
		err = download(torrent, watchDir) // I know I'm shadowing err
		if err == nil {
			torrent.Episode.Pending = false
		}
	}

	return err
}

// This is an odd function here. Perhaps I'll group it with the 'getBody' function.
func download(torrent search_engines.Torrent, watchDir string) error {
	//fileName := torrent.Episode.AsFileName() + ".torrent"

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(path.Join(watchDir, torrent.OriginalName))
	if err != nil {
		return err
	}
	defer output.Close()

	response, err := http.Get(torrent.URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func SearchTorrents(episodes []*sources.Episode) ([]search_engines.Torrent, error) {
	fmt.Printf("Searching for %d torrents", len(episodes))
	c := startProgressBar()
	defer stopProgressBar(c)

	return search_engines.Search(episodes)
}

func Search(query string) ([]sources.Match, []error) {
	fmt.Print("Seaching: ")
	fmt.Print(strings.Join(sources.ListSources(), ", "))
	fmt.Print("\n")

	c := startProgressBar()
	defer stopProgressBar(c)

	matches, errors := sources.Search(query)
	if !isAllNil(errors) {
		return nil, errors
	}

	return matches, nil
}

func isAllNil(errors []error) bool {
	for _, e := range errors {
		if e != nil {
			return false
		}
	}
	return true
}

func Lookup(m *sources.Show) error {
	fmt.Print("Looking up seasons and episodes")
	c := startProgressBar()
	defer stopProgressBar(c)

	return m.GetSeasonsAndEpisodes()
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
		fmt.Printf("err %+v\n", err)
	}
	return strings.Trim(line, "\n")
}
