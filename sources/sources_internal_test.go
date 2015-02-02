package sources

//func TestUpdateSeasonsAndEpisodes(t *testing.T) {
//stack := []string{
//"testdata/updated_seasons.json", // Fixture contains 1 new episode in season 2 and 1 new season.
//"testdata/updated_season_1.json",
//"testdata/updated_season_2.json",
//"testdata/updated_season_3.json",
//}
//var f string

//ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//w.Header().Set("Content-Type", "application/json")
//f, stack = stack[0], stack[1:len(stack)] // *POP*
//fmt.Fprintln(w, readFixture(f))
//}))
//defer ts.Close()

//traktURL = ts.URL

//season1 := &store.Season{Season: 1, Episodes: []*store.Episode{
//{Episode: 1},
//{Episode: 2}}}
//season2 := &store.Season{Season: 2, Episodes: []*store.Episode{
//{Episode: 1},
//{Episode: 2}}}
//s := store.Show{SourceName: traktName, Seasons: []*store.Season{season1, season2}}

//err := UpdateSeasonsAndEpisodes(&s)
//if err != nil {
//t.Fatal("Expected not an error, got:", err)
//}

//if len(s.Seasons) != 3 {
//t.Error("Expected 3 seasons (1 new), got:", len(s.Seasons))
//}
//if len(s.Episodes()) != 7 {
//t.Error("Expected 7 episodes (3 new), got:", len(s.Episodes()))
//}
//}
