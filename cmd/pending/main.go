package main

import (
	"fmt"
	"os"

	"github.com/haarts/getme/config"
	"github.com/haarts/getme/store"
)

func main() {
	store, err := store.Open(config.Config().StateDir)
	if err != nil {
		fmt.Println("Error opening state")
		os.Exit(1)
	}
	defer store.Close()

	for _, v := range store.Shows() {
		fmt.Printf("Show: %s\n", v.Title)
		fmt.Printf("Pending seasons: ")
		for _, season := range v.PendingSeasons() {
			fmt.Printf("%d, ", season.Season)
		}
		fmt.Println("")
		fmt.Println("Pending episodes:")
		for _, episode := range v.PendingEpisodes() {
			fmt.Printf(
				"%02dx%02d - %s\n",
				episode.Season(),
				episode.Episode,
				episode.Title,
			)
		}
		fmt.Println("")
	}
}
