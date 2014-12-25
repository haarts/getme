package sources_test

import (
	"testing"

	"github.com/haarts/getme/sources"
)

func searchTestFunction(_ string) ([]sources.Match, error) {
	return make([]sources.Match, 0), nil
}

func TestAllEpisodesPending(t *testing.T) {
	episodes := []*sources.Episode{
		{Pending: true},
		{Pending: false},
	}
	season := sources.Season{Episodes: episodes}
	if season.AllEpisodesPending() {
		t.Error("Not all episodes are pending")
	}

	episodes = []*sources.Episode{
		{Pending: true},
		{Pending: true},
	}
	season = sources.Season{Episodes: episodes}
	if !season.AllEpisodesPending() {
		t.Error("All episodes are pending")
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
	sources.Register("one", searchTestFunction)
	sources.Register("one", searchTestFunction)
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
