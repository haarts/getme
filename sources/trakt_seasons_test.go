package sources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExpandShow(t *testing.T) {
	stack := []string{
		"fixtures/trakt/seasons.json",
		"fixtures/trakt/season_0.json",
		"fixtures/trakt/season_1.json",
		"fixtures/trakt/season_2.json",
		"fixtures/trakt/season_3.json",
		"fixtures/trakt/season_4.json",
		"fixtures/trakt/season_5.json",
	}

	var f string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f, stack = stack[0], stack[1:len(stack)] // *POP*
		fmt.Fprintln(w, readFixture(f))
	}))
	defer ts.Close()

	traktURL = ts.URL

	show := &Show{URL: "boo/some-url", SourceName: traktName, Title: "Awesome"}
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
	if len(season.Episodes) != 10 {
		t.Fatal("Expected 10 episodes, got:", len(show.Seasons[0].Episodes))
	}

	episode := season.Episodes[0]
	if episode.Episode != 1 {
		t.Error(
			"Expected episode number to equal 10, got:",
			episode.Episode,
		)
	}
}
