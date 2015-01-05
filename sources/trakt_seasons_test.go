package sources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExpandShow(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("fixtures/trakt_seasons.json"))
	}))
	defer ts.Close()

	traktSeasonsURL = ts.URL + "/shows/%s/seasons?extended=full"

	show := &Show{URL: "boo/some-url", SourceName: TRAKT, Title: "Awesome"}
	err := GetSeasonsAndEpisodes(show)
	if err != nil {
		t.Fatal("Expected not an error, got:", err)
	}
	if len(show.Seasons) != 6 {
		t.Fatal("Expected 6 seasons, got:", len(show.Seasons))
	}

	if show.Seasons[1].Season == 0 {
		t.Error("Expected Season field to be not default, got:", show.Seasons[1])
	}

	season := show.Seasons[0]
	if season.Show != show {
		t.Error("Expect Show to point to parent Show, got:", season.Show)
	}
	if len(season.Episodes) != 10 {
		t.Fatal("Expected 10 episodes, got:", len(show.Seasons[0].Episodes))
	}

	episode := season.Episodes[0]
	fmt.Printf("show.Seasons[0] %p\n", show.Seasons[0])
	fmt.Printf("episode.Season %p\n", episode.Season)
	if show.Seasons[0] != episode.Season {
		t.Error(
			"Expected episode to point to parent Season, got:",
			episode.Season,
			"and",
			show.Seasons[0],
		)
	}
}
