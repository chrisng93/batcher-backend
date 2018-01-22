package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadFromURL(url string, song songModel) error {
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
