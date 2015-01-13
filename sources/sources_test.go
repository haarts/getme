package sources_test

import (
	"testing"

	"github.com/haarts/getme/sources"
)

func TestPendingItems(t *testing.T) {
	show := sources.Show{}
	episodes := []*sources.Episode{
		{Pending: true},
		{Pending: true},
		{Pending: true},
	}
	season1 := sources.Season{Season: 1, Episodes: episodes}
	show.Seasons = append(show.Seasons, &season1)
	if len(show.PendingSeasons()) != 0 {
		t.Error("All episodes are pending but it's from the last seasons thus no seasons should be returned, got:", len(show.PendingSeasons()))
	}
	if len(show.PendingEpisodes()) != 3 {
		t.Error("All episodes are pending, got:", len(show.PendingEpisodes()))
	}

	episodes = []*sources.Episode{
		{Pending: true},
		{Pending: true},
	}
	season2 := sources.Season{Season: 2, Episodes: episodes}
	show.Seasons = append(show.Seasons, &season2)
	if len(show.PendingSeasons()) != 1 {
		t.Error("Expected 2 items representing the episodes of the last season and 1 item representing the first season.")
	}
	if len(show.PendingEpisodes()) != 2 {
		t.Error("Expected 2 items representing the episodes of the last season and 1 item representing the first season.")
	}
}

func TestRegisterDuplicateSource(t *testing.T) {
	defer func() {
		str := recover()
		if str == nil {
			t.Error("Expected panic, got none.")
		}
	}()
	sources.Register("one", sources.Trakt{})
	sources.Register("one", sources.Trakt{})
}

func TestDisplayTitle(t *testing.T) {
	s := sources.Show{Title: "bar"}
	if s.DisplayTitle() != s.Title {
		t.Error("Expected DisplayTitle to return the Title, got: ", s.DisplayTitle())
	}
}

func TestEpisodes(t *testing.T) {
	season1 := &sources.Season{Episodes: []*sources.Episode{
		{Episode: 1},
		{Episode: 2}}}
	season2 := &sources.Season{Episodes: []*sources.Episode{
		{Episode: 1},
		{Episode: 2}}}
	s := sources.Show{Seasons: []*sources.Season{season1, season2}}

	if len(s.Episodes()) != 4 {
		t.Error("Expected to have 4 episodes, got: ", len(s.Episodes()))
	}
}
