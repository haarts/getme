package torrents

type TorrentCD struct{}

var torrentCDURL = "http://torrent.cd"

func (t TorrentCD) Name() string {
	return "torrentCD"
}

func (t TorrentCD) Search(query string) ([]Torrent, error) {
	return nil, nil
}

type torrentCDSearchResult struct {
	Channel struct {
		Items []torrentCDItem `xml:"item"`
	} `xml:"channel"`
}

type torrentCDItem struct {
}
