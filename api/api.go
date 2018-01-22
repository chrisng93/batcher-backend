package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	finishedDownloads := make(chan bool)
	for _, song := range body["songs"] {
		go func(song download.Song) {
			defer wg.Done()
			log.Printf("Getting download URL for song. Artist: %v, title: %v", song.Artist, song.Title)
			downloadURL := download.GetDownloadURL(song)
			log.Printf("Downloading song. Artist: %v, title: %v", song.Artist, song.Title)
			err := downloadFromURL(downloadURL, song)
			if err != nil {
				log.Print(err)
			}
			log.Printf("Downloaded song. Artist: %v, title: %v", song.Artist, song.Title)
			finishedDownloads <- true
		}(song)
	}
	go func() {
		wg.Wait()
		fmt.Println("closed download channel")
		close(finishedDownloads)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte{})
}

func downloadFromURL(url string, song download.Song) error {
	fileName := fmt.Sprintf("./downloads/%v - %v.mp3", song.Artist, song.Title)
	output, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating filename %v: %v", fileName, err)
	}
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error getting response for url %v: %v", url, err)
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return fmt.Errorf("error downloading body for url %v: %v", url, err)
	}

	return nil
}
