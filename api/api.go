package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Init initializes the router and adds a handler function for the song route.
func Init() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/song/download", downloadSongsHandler).Methods("POST", "OPTIONS")
	return r
}

func downloadSongsHandler(w http.ResponseWriter, r *http.Request) {
	body := map[string][]songModel{"songs": []songModel{}}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Println(body["songs"])

	var wg sync.WaitGroup
	wg.Add(len(body["songs"]))
	successfulChan := make(chan songModel, len(body["songs"]))
	unsuccessfulChan := make(chan songModel, len(body["songs"]))

	var songs []songModel
	for _, song := range body["songs"] {
		songs = append(songs, song)
		go func(song songModel) {
			defer wg.Done()
			log.Printf("Getting download URL for %v by %v", song.Title, song.Artist)
			downloadURL := getDownloadURL(song, 1)
			if downloadURL == "" {
				log.Printf("Error downloading %v by %v", song.Title, song.Artist)
				unsuccessfulChan <- song
				return
			}

			log.Printf("Downloading %v by %v", song.Title, song.Artist)
			err := downloadFromURL(downloadURL, song)
			if err != nil {
				log.Print(err)
				unsuccessfulChan <- song
			} else {
				log.Printf("Downloaded %v by %v", song.Title, song.Artist)
				successfulChan <- song
			}
			fmt.Println("returning")
			return
		}(song)
	}

	go func() {
		wg.Wait()

		log.Println("======================== REPORT ========================")
		log.Println("")
		log.Println("Successful Downloads:")
		printDownloads(successfulChan)
		fmt.Println("")
		log.Println("Unsuccessful Downloads:")
		printDownloads(unsuccessfulChan)
		log.Println("")
		log.Println("========================================================")
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte{})
}

func printDownloads(songChannel chan songModel) {
	fmt.Println(fmt.Sprintf("Count: %v", len(songChannel)))
	for song := range songChannel {
		fmt.Println(fmt.Sprintf("%v by %v", song.Artist, song.Title))
	}
}
