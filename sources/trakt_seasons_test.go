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

	traktSeasonsURL = ts.URL + "/"

	show := &Show{URL: "boo/some-url", SourceName: TRAKT}
	GetSeasonsAndEpisodes(show)
	if len(show.Seasons) != 6 {
		t.Fatal("Expected 6 seasons, got:", len(show.Seasons))
	}

	if show.Seasons[0].Season == 0 {
		t.Error("Expected Season field to be not default, got:", show.Seasons[0])
	}

	season := show.Seasons[0]
	if season.Show != show {
		t.Error("Expect Show to point to parent Show, got:", season.Show)
	}
	if len(season.Episodes) != 9 {
		t.Fatal("Expected 9 episodes, got:", len(show.Seasons[0].Episodes))
	}

	episode := show.Seasons[0].Episodes[0]
	if show.Seasons[0] != episode.Season {
		t.Error("Expected episode to point to parent Season, got:", episode.Season, show.Seasons[0])
	}
}
