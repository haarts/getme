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

func displayBestMatchConfirmation() bool {
	fmt.Print("Is this the one you want? [Y/n] ")

	bio := bufio.NewReader(os.Stdin)
	line, err := bio.ReadString('\n')
	if err != nil {
		fmt.Printf("err %+v\n", err)
	}

	if line == "\n" || line == "y\n" || line == "Y\n" {
		return true
	} else {
		return false
	}
}

func displayAlternatives(ms []sources.Match) *sources.Match {
	fmt.Println("Which one ARE you looking for?")
	for i, m := range ms {
		fmt.Printf("[%d] %s\n", i+1, m.Title)
	}

	fmt.Print("Enter the correct number: ")

	bio := bufio.NewReader(os.Stdin)
	line, err := bio.ReadString('\n')
	if err != nil {
		fmt.Printf("err %+v\n", err)
	}

	// User abort
	if line == "\n" {
		return nil
	}

	withoutNewline := strings.Trim(line, "\n")
	i, err := strconv.Atoi(withoutNewline)
	// User mis-typed, try again
	if err != nil {
		return displayAlternatives(ms)
	}

	return &ms[i-1]
}

func search(query string) []sources.Match {
	c := time.NewTicker(1 * time.Second)
	go func() {
		for _ = range c.C {
			fmt.Print(".")
		}
	}()

	matches, err := sources.Search(query)
	if err != nil {
		fmt.Printf("err %+v\n", err)
	}
	c.Stop()
	fmt.Print("\n")
	return matches
}

func main() {
	store := store.Open()
	query := getQuery()
	matches := search(query)
	if len(matches) == 0 {
		fmt.Println("We haven't found what you were looking for.")
		return
	}

	displayBestMatch(matches[0])
	bestMatchConfirmed := displayBestMatchConfirmation()
	if bestMatchConfirmed {
		store.CreateShow(matches[0])
	} else {
		match := displayAlternatives(matches)
		if match != nil {
			store.CreateShow(match)
		}
	}
	fmt.Printf("store %+v\n", store)
}
