package torrents_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/haarts/getme/torrents"
)

func TestSearch(t *testing.T) {
	assert.Fail(t, "implement me")
}

func TestIsEnglish(t *testing.T) {
	ss := []string{
		"it's all good",
		"this is very french",
		"some show vostfr",
		"some.show.ITA.avi",
	}

	assert.True(t, torrents.IsEnglish(ss[0]))

	for _, s := range ss[1:] {
		assert.False(t, torrents.IsEnglish(s), "should not be english: %s", s)
	}
}
