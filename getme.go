package main

import (
	"bufio"
	"fmt"
	"os"

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
	fmt.Println(" ", bestMatch.Title())
}

func askConfirmation() bool {
	fmt.Print("Is this the one you want? [Y/n]")
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

func main() {
	query := getQuery()
	matches, err := sources.Search(query)
	if err != nil {
		fmt.Printf("err %+v\n", err)
	}

	showBestMatch(matches.BestMatch())
	bestMatchConfirmed := askConfirmation()
	fmt.Printf("bestMatchConfirmed %+v\n", bestMatchConfirmed)
	if bestMatchConfirmed {
		// Store it somewhere.
	} else {
		// Show entire list and store selection.
	}

	fmt.Printf("matches %+v\n", matches)
	fmt.Printf("matches.BestMatch() %+v\n", matches.BestMatch())
}
