package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Song defines the structure of a single song input.
type Song struct {
	URL    string `json:"permalink_url"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

// Init initializes the router and adds a handler function for the song route.
func Init() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/song/download", downloadSongsHandler).Methods("POST")
	return r
}

func downloadSongsHandler(w http.ResponseWriter, r *http.Request) {
	body := map[string][]Song{"songs": []Song{}}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Println(body["songs"])

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte{})
}
