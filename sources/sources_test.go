package sources_test

import (
	"testing"

	"github.com/haarts/getme/sources"
)

func TestRegisterNilSource(t *testing.T) {
	defer func() {
		str := recover()
		if str == nil {
			t.Error("Expected panic, got none.")
		}
	}()
	sources.Register("nil", nil)
}

func TestRegisterDuplicateSource(t *testing.T) {
	defer func() {
		str := recover()
		if str == nil {
			t.Error("Expected panic, got none.")
		}
	}()
	sources.Register("one", 123)
	sources.Register("one", 123)
}

// TODO Not sure how to test this yet.
func TestRegister(t *testing.T) {
	sources.Register("one", 123)
}

func TestCreateEpisodes(t *testing.T) {
	seasons := []sources.Season{{"bar", 1, 4}, {"bar", 2, 5}}

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
