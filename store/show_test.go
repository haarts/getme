package store_test

import (
	"testing"
	"time"

	"github.com/haarts/getme/store"
	"github.com/stretchr/testify/assert"
)

func TestDetermineDailyUnhappy(t *testing.T) {
	assert.NotPanics(t, func() {
		(&store.Show{}).DetermineIsDaily()
	})
}

func TestDetermineDaily(t *testing.T) {
	initDate, _ := time.Parse("2006-01-02", "1996-07-22")
	nextDay, _ := time.Parse("2006-01-02", "1996-07-23")
	episodes := make([]*store.Episode, 31) // minimum length
	episodes[4] = &store.Episode{AirDate: initDate, Episode: 1}
	episodes[5] = &store.Episode{AirDate: nextDay, Episode: 2}
	season := &store.Season{Episodes: episodes}

	show := store.Show{Seasons: []*store.Season{season}}
	if !show.DetermineIsDaily() {
		t.Error("Expected the show to be daily.")
	}

	initDate, _ = time.Parse("2006-01-02", "1996-07-22")
	daysLater, _ := time.Parse("2006-01-02", "1996-07-24")
	episodes[4] = &store.Episode{AirDate: initDate, Episode: 1}
	episodes[5] = &store.Episode{AirDate: daysLater, Episode: 2}
	season = &store.Season{Episodes: episodes}

	show = store.Show{Seasons: []*store.Season{season}}
	if show.DetermineIsDaily() {
		t.Error("Expected the show to be NOT daily.")
	}
}
