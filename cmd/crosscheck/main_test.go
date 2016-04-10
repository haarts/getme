package main

import (
	"os"
	"testing"
	"time"

	"github.com/haarts/getme/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFileInfo struct {
	name  string
	isDir bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 1 }
func (m mockFileInfo) Mode() os.FileMode  { return os.ModeDir }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

func TestMatchDirWithShow(t *testing.T) {
	backlessStore := &store.Store{}

	show1 := &store.Show{Title: "foo"}
	require.NoError(t, backlessStore.CreateShow(show1))

	show2 := &store.Show{Title: "bar"}
	require.NoError(t, backlessStore.CreateShow(show2))

	isShow := mockFileInfo{name: "foo", isDir: true}
	isNotDir := mockFileInfo{name: "foo", isDir: false}
	isNotShow := mockFileInfo{name: "baz", isDir: true}

	assert.Equal(t, show1, matchDirWithShow(isShow, backlessStore))
	assert.Nil(t, matchDirWithShow(isNotDir, backlessStore))
	assert.Nil(t, matchDirWithShow(isNotShow, backlessStore))
}

func TestVerifyPendingStates(t *testing.T) {
	show := &store.Show{
		Title: "foo",
		Seasons: []*store.Season{
			&store.Season{
				Season: 1,
				Episodes: []*store.Episode{
					&store.Episode{
						Episode: 1,
						Pending: false,
					},
					&store.Episode{
						Episode: 2,
						Pending: false,
					},
				},
			},
		},
	}

	require.False(t, show.Seasons[0].Episodes[0].Pending)
	require.False(t, show.Seasons[0].Episodes[1].Pending)

	verifyPendingStates(mockFileInfo{name: "testdata/Videos/foo"}, show)
	assert.True(t, show.Seasons[0].Episodes[0].Pending)
	assert.False(t, show.Seasons[0].Episodes[1].Pending)
}
