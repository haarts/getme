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
	if len(show.PendingItems()) != 3 {
		t.Error("All episodes are pending")
	}

	episodes = []*sources.Episode{
		{Pending: true},
		{Pending: true},
	}
	season2 := sources.Season{Season: 2, Episodes: episodes}
	show.Seasons = append(show.Seasons, &season2)
	if len(show.PendingItems()) != 3 {
		t.Error("Expected 2 items representing the episodes of the last season and 1 item representing the first season.")
	}
}

func TestAsFileName(t *testing.T) {
	show := sources.Show{Title: "with & silly ! chars 123"}
	season := sources.Season{Show: &show}
	episode := sources.Episode{Season: &season}

	if episode.AsFileName() != "with___silly___chars_123_S00E00" {
		t.Error("Expected no silly characters, got:", episode.AsFileName())
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
