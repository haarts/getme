package main

import (
	"fmt"
	"os"

	"github.com/haarts/getme/providers"
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
	matches := providers.Search(query)
	fmt.Printf("matches %+v\n", matches)
}
