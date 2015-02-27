package sources

import (
	"testing"

	"github.com/haarts/getme/store"
	"github.com/stretchr/testify/assert"
)

func TestReplaceTBAEpisode(t *testing.T) {
	existingEpisodes := []*store.Episode{
		{Episode: 1, Title: "Regular"},
		{Episode: 2, Title: "TBA"},
	}
	existingSeason := store.Season{Episodes: existingEpisodes}

	present := Episode{Episode: 1, Title: "Not get replaced"}
	replacement := Episode{Episode: 2, Title: "Second one"}
	newEpisode := Episode{Episode: 3, Title: "Third one"}

	newSeason := Season{Episodes: []Episode{present, replacement, newEpisode}}

	updateEpisodes(&existingSeason, newSeason)

	assert.Equal(t, replacement.Title, existingSeason.Episodes[1].Title)
}
