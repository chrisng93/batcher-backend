package download

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
)

const converterURL = "http://convert2mp3.net/en/"

// Song defines the structure of a single song input.
type Song struct {
	URL    string `json:"permalink_url"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

// GetDownloadURL gets and returns the download URL for a song.
func GetDownloadURL(song Song) string {
	// Set context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create chrome instance.
	var options chromedp.Option
	options = chromedp.WithRunnerOptions(
		runner.Flag("headless", true),
		runner.Flag("no-sandbox", true),
	)
	c, err := chromedp.New(ctx, options, chromedp.WithErrorf(log.Printf))
	if err != nil {
		log.Fatalf("Error creating chrome instance: %v", err)
	}

	// Get the song download URL.
	var downloadURL string
	err = getURL(ctx, c, song, &downloadURL)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("got download url", downloadURL)

	err = c.Shutdown(ctx)
	if err != nil {
		log.Printf("Error shutting down chrome instance: %v", err)
	}

	err = c.Wait()
	if err != nil {
		log.Printf("Error waiting for chrome instance to finish: %v", err)
	}

	return downloadURL
}

func getURL(ctx context.Context, c *chromedp.CDP, song Song, downloadURL *string) error {
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	if err := c.Run(ctx, chromedp.Navigate(converterURL)); err != nil {
		return fmt.Errorf("error navigating to converter URL: %v", err)
	}

	// Set song URL.
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible(".input_convert", chromedp.ByQuery),
		chromedp.SetValue(".input_convert", song.URL, chromedp.ByQuery),
		chromedp.Click("#convertForm .mainbtn", chromedp.NodeVisible),
		chromedp.Sleep(3 * time.Second),
	}); err != nil {
		return fmt.Errorf("error setting song URL: %v", err)
	}

	// Set song artist.
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible("#input_artist", chromedp.ByID),
		chromedp.Click("#input_artist a", chromedp.NodeVisible),
		chromedp.WaitVisible("#input_artist input", chromedp.ByQuery),
		chromedp.SetValue("#input_artist input", song.Artist, chromedp.ByQuery),
	}); err != nil {
		return fmt.Errorf("error setting song artst: %v", err)
	}

	// Set song title.
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible("#input_title", chromedp.ByID),
		chromedp.Click("#input_title a", chromedp.NodeVisible),
		chromedp.WaitVisible("#input_title input", chromedp.ByQuery),
		chromedp.SetValue("#input_title input", song.Title, chromedp.ByQuery),
	}); err != nil {
		return fmt.Errorf("error setting song title: %v", err)
	}

	// Set album cover.
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible("#advancedtagsbtn", chromedp.ByID),
		chromedp.Click("#advancedtagsbtn a", chromedp.NodeVisible),
		chromedp.WaitVisible("#inputCover", chromedp.ByID),
		chromedp.Click("#inputCover", chromedp.NodeVisible),
		chromedp.Sleep(3 * time.Second),
		chromedp.Click(".btn-success", chromedp.NodeVisible),
		chromedp.Sleep(1 * time.Second),
	}); err != nil {
		return fmt.Errorf("error setting album cover: %v", err)
	}

	// Get song download URL.
	var downloadOK bool
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible(".btn-success", chromedp.ByQuery),
		chromedp.AttributeValue(".btn-success", "href", downloadURL, &downloadOK, chromedp.ByQuery),
	}); err != nil {
		return fmt.Errorf("error getting song download URL: %v", err)
	}

	return nil
}
