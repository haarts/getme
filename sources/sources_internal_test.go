package sources

import (
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
