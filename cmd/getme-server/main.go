package main

import (
	"net/http"

	"github.com/haarts/getme/config"
	"github.com/haarts/getme/store"
)

type searchHandler struct {
	store *store.Store
	log   *config.Logger
}

func (h *searchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "parameter 'q' required", http.StatusBadRequest)
		return
	}
}

func main() {
	config := config.Config()
	store, _ := store.Open(config.StateDir)
	log := config.Logger
	http.Handle("/search", &searchHandler{store, log})

	http.ListenAndServe(":1999", nil)
}
