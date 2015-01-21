package sources

import (
	"testing"

	"github.com/haarts/getme/store"
)

func TestPopularShowAtIndex(t *testing.T) {
	oldShows := SHOWS
	SHOWS = [100]string{
		"one",
		"two",
		"three",
	}

	matches := []Match{
		store.Show{Title: "not one"},
		store.Show{Title: "two"},
	}

	if popularShowAtIndex(matches) != 1 {
		t.Error("Expected to find the popular show at index 1, got:", popularShowAtIndex(matches))
	}
	SHOWS = oldShows // reset
}
