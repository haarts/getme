package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/haarts/getme/search_engines"
	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
	"github.com/haarts/getme/ui"
)

func handleShow(show *sources.Show) error {
	store := store.Open()
	defer store.Close()

	// Fetch the seasons/episodes associated with the found show.
	err := ui.Lookup(show)
	if err != nil {
		fmt.Println("We've encountered a problem looking up seasons for the show. The error:")
		fmt.Println(" ", err)
		return err
	}

	store.CreateShow(*show)

	// We have two entry points. One on the first run and one when running as daemon.
	// So we create episodes based on seasons always. Then look at the disk/store and figure out
	// what is missing.

	torrents, err := ui.SearchTorrents(show.Episodes())
	if err != nil {
		fmt.Println("Something went wrong looking for your torrents: ", err)
		return err
	}
	for _, torrent := range torrents {
		err := download(torrent)
		if err != nil {
			fmt.Println("Something went wrong downloading a torrent: ", err)
		}
	}

	return nil
}

func main() {
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
}

// This is an odd function here. Perhaps I'll group it with the 'getBody' function.
func download(torrent search_engines.Torrent) error {
	fileName := torrent.Episode.AsFileName()
	fmt.Println("Downloading", torrent.URL, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create("/tmp/getme/" + fileName)
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
