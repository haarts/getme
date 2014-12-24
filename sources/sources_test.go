package sources_test

import (
	"testing"

	"github.com/haarts/getme/sources"
)

func searchTestFunction(_ string) ([]sources.Match, error) {
	return make([]sources.Match, 0), nil
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
