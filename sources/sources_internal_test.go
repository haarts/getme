package sources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsDaily(t *testing.T) {
	initDate, _ := time.Parse("2006-01-02", "1996-07-22")
	nextDay, _ := time.Parse("2006-01-02", "1996-07-23")
	episodes := make([]*Episode, 31) // minimum length
	episodes[4] = &Episode{AirDate: initDate, Episode: 1}
	episodes[5] = &Episode{AirDate: nextDay, Episode: 2}
	season := &Season{Episodes: episodes}

	show := Show{Seasons: []*Season{season}}
	if !show.determineIsDaily() {
		t.Error("Expected the show to be daily.")
	}

	initDate, _ = time.Parse("2006-01-02", "1996-07-22")
	daysLater, _ := time.Parse("2006-01-02", "1996-07-24")
	episodes[4] = &Episode{AirDate: initDate, Episode: 1}
	episodes[5] = &Episode{AirDate: daysLater, Episode: 2}
	season = &Season{Episodes: episodes}

	show = Show{Seasons: []*Season{season}}
	if show.determineIsDaily() {
		t.Error("Expected the show to be NOT daily.")
	}
}

func TestUpdateSeasonsAndEpisodes(t *testing.T) {
	stack := []string{
		"fixtures/updated_seasons.json", // Fixture contains 1 new episode in season 2 and 1 new season.
		"fixtures/updated_season_1.json",
		"fixtures/updated_season_2.json",
		"fixtures/updated_season_3.json",
	}
	var f string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f, stack = stack[0], stack[1:len(stack)] // *POP*
		fmt.Fprintln(w, readFixture(f))
	}))
	defer ts.Close()

	traktURL = ts.URL

	season1 := &Season{Season: 1, Episodes: []*Episode{
		{Episode: 1},
		{Episode: 2}}}
	season2 := &Season{Season: 2, Episodes: []*Episode{
		{Episode: 1},
		{Episode: 2}}}
	s := Show{SourceName: traktName, Seasons: []*Season{season1, season2}}

	err := UpdateSeasonsAndEpisodes(&s)
	if err != nil {
		t.Fatal("Expected not an error, got:", err)
	}

	if len(s.Seasons) != 3 {
		t.Error("Expected 3 seasons (1 new), got:", len(s.Seasons))
	}
	if len(s.Episodes()) != 7 {
		t.Error("Expected 7 episodes (3 new), got:", len(s.Episodes()))
	}
}
