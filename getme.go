package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/haarts/getme/sources"
)

func getQuery() string {
	if len(os.Args) != 2 {
		fmt.Println("Please pass a search query.")
		os.Exit(1)
	}

	query := os.Args[1]
	return query
}

func showBestMatch(bestMatch sources.Match) {
	fmt.Println("The best match we found is:")
	fmt.Println(" ", bestMatch.Title)
}

func askConfirmation() bool {
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

func presentAlternatives(ms []sources.Match) *sources.Match {
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

	if line == "\n" {
		return nil
	}

	withoutNewline := strings.Trim(line, "\n")
	i, err := strconv.Atoi(withoutNewline)
	if err != nil {
		return presentAlternatives(ms)
	}

	return &ms[i-1]
}

func main() {
	query := getQuery()
	matches, err := sources.Search(query)
	if err != nil {
		fmt.Printf("err %+v\n", err)
	}

	showBestMatch(matches[0])
	bestMatchConfirmed := askConfirmation()
	if bestMatchConfirmed {
		// Store it somewhere.
	} else {
		match := presentAlternatives(matches)
		if match != nil {
			// Store it somewhere.
		}
	}
}
