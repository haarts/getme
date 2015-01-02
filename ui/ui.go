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
	"text/tabwriter"
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

func DisplayBestMatchConfirmation(matches [][]sources.Match) *sources.Match {
	nonNilMatch := firstNonNilMatch(matches)
	if nonNilMatch == nil {
		return nil
	}

	displayBestMatch(*nonNilMatch)
	fmt.Print("Is this the one you want? [Y/n] ")
	line := getUserInput()

	if line == "" || line == "y" || line == "Y" {
		return nonNilMatch
	} else {
		return nil
	}
}

func firstNonNilMatch(matches [][]sources.Match) *sources.Match {
	for _, ms := range matches {
		if len(ms) != 0 {
			return &ms[0]
		}
	}
	return nil // This really shouldn't happen.
}

// TODO break this func up. Too long.
func DisplayAlternatives(ms [][]sources.Match) *sources.Match {
	fmt.Println("Which one ARE you looking for?")
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprint(w, strings.Join(sources.ListSources(), "\t")+"\n")

	var generators []func() (string, []interface{})
	step := 1
	for i, m := range ms {
		if i > 0 {
			step += len(ms[i-1])
		}
		generators = append(generators, createGenerator(m, step))
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
		return DisplayAlternatives(ms)
	}

	var flatList []sources.Match
	for _, m := range ms {
		flatList = append(flatList, m...)
	}

	return &flatList[i-1]
}

func createGenerator(ms []sources.Match, step int) func() (string, []interface{}) {
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

func Search(query string) ([][]sources.Match, []error) {
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

func Lookup(s *sources.Show) error {
	fmt.Print("Looking up seasons and episodes for ", s.Title)
	c := startProgressBar()
	defer stopProgressBar(c)

	return sources.GetSeasonsAndEpisodes(s)
	//return s.GetSeasonsAndEpisodes()
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
