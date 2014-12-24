package sources

import "testing"

func TestPopularShowAtIndex(t *testing.T) {
	shows = [100]string{
		"one",
		"two",
		"three",
	}

	matches := []Match{
		Show{Title: "not one"},
		Show{Title: "two"},
	}

	if popularShowAtIndex(matches) != 1 {
		t.Error("Expected to find the popular show at index 1, got:", popularShowAtIndex(matches))
	}
}
