package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

func DisplayAlternatives(ms []sources.Match) *sources.Match {
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
		return DisplayAlternatives(ms)
	}

	return &ms[i-1]
}

func Search(query string) []sources.Match {
	fmt.Print("Seaching")
	c := startProgressBar()
	defer stopProgressBar(c)

	matches, err := sources.Search(query)
	if err != nil { // TODO: Handle the error upstream
		fmt.Printf("err %+v\n", err)
	}

	return matches
}

func SearchSeasons(m sources.Match) ([]sources.Season, error) {
	fmt.Print("Looking up seasons")
	c := startProgressBar()
	defer stopProgressBar(c)

	seasons, err := sources.GetSeasons(m)
	if err != nil {
		return nil, err
	}

	return seasons, nil
}

func displayBestMatch(bestMatch sources.Match) {
	fmt.Println("The best match we found is:")
	fmt.Println(" ", bestMatch.Title)
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
