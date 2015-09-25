package torrents

import "github.com/haarts/getme/store"

type ExtraTorrent struct{}

func (t ExtraTorrent) Search(show *store.Show) ([]Torrent, error) {
	return nil, nil
}
