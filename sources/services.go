package sources

func popularShowAtIndex(shows []Match) int {
	for i, m := range shows {
		if isPopularShow(m.DisplayTitle()) {
			return i
		}
	}

	return -1
}

func isPopularShow(showName string) bool {
	for _, n := range SHOWS {
		if showName == n {
			return true
		}
	}
	return false
}
