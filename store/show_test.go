package store_test

import (
	"sort"
	"testing"
	"time"

	"github.com/haarts/getme/store"
	"github.com/stretchr/testify/assert"
)

func TestStoreEpisodeSnippet(t *testing.T) {
	show := store.Show{}
	snippet := store.Snippet{Score: 123, TitleSnippet: "abc", FormatSnippet: "sxs"}
	show.StoreEpisodeSnippet(snippet)

	assert.Equal(t, snippet, show.QuerySnippets.ForEpisode[0])
}

func TestOverWriteStoreEpisodeSnippet(t *testing.T) {
	show := store.Show{}

	snippet := store.Snippet{Score: 123, TitleSnippet: "abc", FormatSnippet: "sxs"}
	show.StoreEpisodeSnippet(snippet)

	snippet = store.Snippet{Score: 234, TitleSnippet: "abc", FormatSnippet: "sxs"}
	show.StoreEpisodeSnippet(snippet)

	assert.Len(t, show.QuerySnippets.ForEpisode, 1)
	assert.Equal(t, snippet, show.QuerySnippets.ForEpisode[0])
}

func TestSortByAirDate(t *testing.T) {
	episodes := []*store.Episode{
		{AirDate: time.Now().Add(-5 * time.Hour), Title: "oldest"},
		{AirDate: time.Now().Add(-1 * time.Hour), Title: "youngest"},
		{AirDate: time.Now().Add(-3 * time.Hour), Title: "middle aged"},
	}

	sort.Sort(store.ByAirDate(episodes))
	if episodes[0].Title != "youngest" {
		t.Error("Expected the younget episode on top, got:", episodes)
	}
}

func TestPendingItems(t *testing.T) {
	show := store.Show{}

	// first and only season is always pending in absence of Ended
	episodesFor1 := []*store.Episode{
		{Pending: true},
		{Pending: true},
		{Pending: true},
	}
	season1 := store.Season{Season: 1, Episodes: episodesFor1}
	show.Seasons = append(show.Seasons, &season1)

	assert.Equal(t, 0, len(show.PendingSeasons()))
	assert.Equal(t, 3, len(show.PendingEpisodes()))

	// unless the show is done
	ended := true
	show.Ended = &ended
	assert.Equal(t, 1, len(show.PendingSeasons()))
	assert.Equal(t, 0, len(show.PendingEpisodes()))

	// but not when it is verifiably running
	ended = false
	assert.Equal(t, 0, len(show.PendingSeasons()))
	assert.Equal(t, 3, len(show.PendingEpisodes()))

	// when there are more than 1 seasons
	show.Ended = nil
	episodesFor2 := []*store.Episode{
		{Pending: true},
		{Pending: true},
	}
	season2 := store.Season{Season: 2, Episodes: episodesFor2}
	show.Seasons = append(show.Seasons, &season2)

	assert.Equal(t, 1, len(show.PendingSeasons()))
	assert.Equal(t, 2, len(show.PendingEpisodes()))
}

func TestDisplayTitle(t *testing.T) {
	s := store.Show{Title: "bar"}
	if s.DisplayTitle() != s.Title {
		t.Error("Expected DisplayTitle to return the Title, got: ", s.DisplayTitle())
	}
}

func TestEpisodes(t *testing.T) {
	season1 := &store.Season{Episodes: []*store.Episode{
		{Episode: 1},
		{Episode: 2}}}
	season2 := &store.Season{Episodes: []*store.Episode{
		{Episode: 1},
		{Episode: 2}}}
	s := store.Show{Seasons: []*store.Season{season1, season2}}

	if len(s.Episodes()) != 4 {
		t.Error("Expected to have 4 episodes, got: ", len(s.Episodes()))
	}
}
