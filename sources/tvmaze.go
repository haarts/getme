package sources

import "github.com/haarts/getme/store"

const tvMazeName = "tvmaze"

type TvMaze struct{}

func (t TvMaze) Name() string {
	return tvMazeName
}

func (t TvMaze) Search(q string) SearchResult {
	return SearchResult{}
}

func (t TvMaze) Seasons(show *store.Show) ([]Season, error) {
	return nil, nil
}
