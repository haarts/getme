package sources

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/haarts/getme/store"
)

func readFixture(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
	return string(data)
}

func TestOrdering(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("testdata/trakt_search.json"))
	}))
	defer ts.Close()
	defer func() {
		sources = make(map[string]Source)
		Register(tvRageName, TvRage{})
		Register(traktName, Trakt{})
	}()

	traktSearchURL = ts.URL + "/search?type=show"

	sources = make(map[string]Source)
	Register(traktName, Trakt{})
	matches := Search("some query")
	if matches[0].Matches[0].DisplayTitle() != "Game of Thrones" {
		t.Fatal("best match is not Game of Thrones, got:", matches[0].Matches[0].DisplayTitle())
	}
}

func TestExpandShow(t *testing.T) {
	stack := []string{
		"testdata/trakt/seasons.json",
		"testdata/trakt/season_0.json",
		"testdata/trakt/season_1.json",
		"testdata/trakt/season_2.json",
		"testdata/trakt/season_3.json",
		"testdata/trakt/season_4.json",
		"testdata/trakt/season_5.json",
	}

	var f string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f, stack = stack[0], stack[1:len(stack)] // *POP*
		fmt.Fprintln(w, readFixture(f))
	}))
	defer ts.Close()

	traktURL = ts.URL

	show := &store.Show{URL: "boo/some-url", SourceName: traktName, Title: "Awesome"}
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
