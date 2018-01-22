package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/chrisng93/batcher-backend/download"
	"github.com/gorilla/mux"
)

// Init initializes the router and adds a handler function for the song route.
func Init() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/song/download", downloadSongsHandler).Methods("POST")
	return r
}

func downloadSongsHandler(w http.ResponseWriter, r *http.Request) {
	body := map[string][]download.Song{"songs": []download.Song{}}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Println(body["songs"])

	var wg sync.WaitGroup
	wg.Add(len(body["songs"]))
	downloadURLs := make(chan string)
	for _, song := range body["songs"] {
		go func(song download.Song) {
			defer wg.Done()
			downloadURLs <- download.GetDownloadURL(song)
		}(song)
	}
	fmt.Println("looped through songs")
	go func() {
		wg.Wait()
		fmt.Println("closed channel")
		close(downloadURLs)
	}()

	var urls []string
	fmt.Println("ranging through download URLS")
	for url := range downloadURLs {
		fmt.Println(url)
		urls = append(urls, url)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte{})
}
