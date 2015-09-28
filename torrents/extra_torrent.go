package torrents

type ExtraTorrent struct{}

func (e ExtraTorrent) Name() string {
	return "extratorrent"
}

func (e ExtraTorrent) Search(query string) ([]Torrent, error) {
	return nil, nil
}
