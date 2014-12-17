package main

import (
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

func main() {
	query := getQuery()
	matches := sources.Search(query)
	fmt.Printf("matches %+v\n", matches)
	fmt.Printf("matches.BestMatch() %+v\n", matches.BestMatch())
}
