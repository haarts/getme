package store

// Movie contains all the relevant information for a movie.
type Movie struct {
	Title string
}

// DisplayTitle returns the title of a movie. Here to satisfy the Match
// interface.
func (m Movie) DisplayTitle() string {
	return m.Title
}
