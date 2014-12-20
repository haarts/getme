package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
)

func getQuery() string {
	if len(os.Args) != 2 {
		fmt.Println("Please pass a search query.")
		os.Exit(1)
	}

	query := os.Args[1]
	return query
}

func displayBestMatch(bestMatch sources.Match) {
	fmt.Println("The best match we found is:")
	fmt.Println(" ", bestMatch.Title)
}

func getUserInput() string {
	bio := bufio.NewReader(os.Stdin)
	line, err := bio.ReadString('\n')
	if err != nil {
		fmt.Printf("err %+v\n", err)
	}
	return strings.Trim(line, "\n")
}

func displayBestMatchConfirmation(matches []sources.Match) *sources.Match {
	displayBestMatch(matches[0])
	fmt.Print("Is this the one you want? [Y/n] ")
	line := getUserInput()

	if line == "" || line == "y" || line == "Y" {
		return &matches[0]
	} else {
		return nil
	}
}

func displayAlternatives(ms []sources.Match) *sources.Match {
	fmt.Println("Which one ARE you looking for?")
	for i, m := range ms {
		fmt.Printf("[%d] %s\n", i+1, m.Title)
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
		return displayAlternatives(ms)
	}

	return &ms[i-1]
}

func search(query string) []sources.Match {
	c := startProgressBar()
	defer stopProgressBar(c)

	matches, err := sources.Search(query)
	if err != nil { // TODO: Handle the error upstream
		fmt.Printf("err %+v\n", err)
	}

	return matches
}

func searchSeasons(m sources.Match) ([]sources.Season, error) {
	c := startProgressBar()
	defer stopProgressBar(c)

	seasons, err := sources.GetSeasons(m)
	if err != nil {
		return nil, err
	}

	return seasons, nil
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

func main() {
	store := store.Open()
	query := getQuery()
	matches := search(query)
	if len(matches) == 0 {
		fmt.Println("We haven't found what you were looking for.")
		return
	}

	// Determine which show ppl want.
	match := displayBestMatchConfirmation(matches)
	if match != nil {
		store.CreateShow(*match)
	} else {
		match := displayAlternatives(matches)
		if match != nil {
			store.CreateShow(*match)
		} else {
			return
		}
	}

	// Fetch the seasons associated with the found show.
	seasons, _ := searchSeasons(*match)
	fmt.Printf("seasons %+v\n", seasons)
}
