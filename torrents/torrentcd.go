package torrents

import "github.com/haarts/getme/store"

type TorrentCD struct{}

var torrentCDURL = "http://torrent.cd"

func (t TorrentCD) Search(show *store.Show) ([]Torrent, error) {
	return nil, nil
}
