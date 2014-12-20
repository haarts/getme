package sources_test

import (
	"testing"

	"github.com/haarts/getme/sources"
)

func TestCreateEpisodes(t *testing.T) {
	seasons := []sources.Season{{1, 4}, {2, 5}}

	episodes := sources.CreateEpisodes(seasons)
	if len(episodes) != 9 {
		t.Error("Expected the total number of episodes to be 9, got: ", len(episodes))
	}

	if episodes[0].Season != 1 {
		t.Error("Expected the first episode to belong to the first season, got: ", episodes[0])
	}

	if episodes[4].Episode != 1 && episodes[4].Season != 2 {
		t.Error("Expected the episode to be of season 2, got: ", episodes[4])
	}
}
