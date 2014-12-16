package main

import (
	"fmt"
	"os"
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
	fmt.Printf("query %+v\n", query)
}
