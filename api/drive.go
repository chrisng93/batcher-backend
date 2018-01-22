package api

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

const converterURL = "http://convert2mp3.net/en/"
const maxTries = 5
const jsCheckForPopup = `
	if (window.location.href.indexOf('convert2mp3') === -1) {
		document.getElementsByTagName('button')[0].click()
		return true;
	}
	return false;
`

// songModel defines the structure of a single song input.
type songModel struct {
	URL    string `json:"permalink_url"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type downloadURLResults struct {
	urls        []string
	successful  []songModel
	unsucessful []songModel
}

// getDownloadURL gets and returns the download URL for a song. On failure, it tries maxTries
// times before marking the download as unsucessful.
func getDownloadURL(song songModel, tries int) string {
	// Set context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create chrome instance.
	var options chromedp.Option
	options = chromedp.WithRunnerOptions(
	// runner.Flag("headless", true),
	// runner.Flag("no-sandbox", true),
	)
	c, err := chromedp.New(ctx, options)
	if err != nil {
		log.Fatalf("Error creating chrome instance: %v", err)
	}

	// Get the song download URL.
	var downloadURL string
	err = getURL(ctx, c, song, &downloadURL)
	if err != nil {
		log.Println(err)
	}

	// Shut down chrome page handlers.
	err = c.Shutdown(ctx)
	if err != nil {
		log.Printf("Error shutting down chrome instance: %v", err)
	}

	// Wait for Chrome runner to terminate.
	err = c.Wait()
	if err != nil {
		log.Printf("Error waiting for chrome instance to finish: %v", err)
	}

	if downloadURL == "" && tries < maxTries {
		log.Printf("Attempt #%v for downloading %v by %v", tries+1, song.Title, song.Artist)
		return getDownloadURL(song, tries+1)
	}
	return downloadURL
}

// getURL walks through the chrome browser to input metadata about and get the download URL for
// the given song.
func getURL(ctx context.Context, c *chromedp.CDP, song songModel, downloadURL *string) error {
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := c.Run(ctx, chromedp.Navigate(converterURL)); err != nil {
		return fmt.Errorf("error navigating to converter URL: %v", err)
	}

	originalTabs := c.ListTargets()

	// Set song URL.
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.Sleep(1 * time.Second),
		chromedp.WaitVisible(".input_convert", chromedp.ByQuery),
		chromedp.SetValue(".input_convert", song.URL, chromedp.ByQuery),
		chromedp.Click("#convertForm .mainbtn", chromedp.NodeVisible),
	}); err != nil {
		return fmt.Errorf("error setting song URL: %v", err)
	}

	updatedTabs := c.ListTargets()
	// Unable to do this right now. Closing tabs is not yet implemented.
	// https://github.com/chromedp/chromedp/issues/144
	if len(originalTabs) != len(updatedTabs) {
		// Delete added tab.
	}

	// Set song artist.
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.Sleep(15 * time.Second),
		chromedp.WaitVisible("#input_artist", chromedp.ByID),
		chromedp.Click("#input_artist a", chromedp.NodeVisible),
		chromedp.WaitVisible("#input_artist input", chromedp.ByQuery),
		chromedp.SetValue("#input_artist input", song.Artist, chromedp.ByQuery),
	}); err != nil {
		return fmt.Errorf("error setting song artst: %v", err)
	}

	// Set song title.
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.Sleep(1 * time.Second),
		chromedp.WaitVisible("#input_title", chromedp.ByID),
		chromedp.Click("#input_title a", chromedp.NodeVisible),
		chromedp.WaitVisible("#input_title input", chromedp.ByQuery),
		chromedp.SetValue("#input_title input", song.Title, chromedp.ByQuery),
	}); err != nil {
		return fmt.Errorf("error setting song title: %v", err)
	}

	// Set album cover.
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.Sleep(1 * time.Second),
		chromedp.WaitVisible("#advancedtagsbtn", chromedp.ByID),
		chromedp.Click("#advancedtagsbtn a", chromedp.NodeVisible),
		chromedp.WaitVisible("#inputCover", chromedp.ByID),
		chromedp.Click("#inputCover", chromedp.NodeVisible),
		chromedp.Click(".btn-success", chromedp.NodeVisible),
	}); err != nil {
		return fmt.Errorf("error setting album cover: %v", err)
	}

	updatedTabs = c.ListTargets()
	if len(originalTabs) != len(updatedTabs) {
		// Delete added tab.
	}

	// Get song download URL.
	var downloadOK bool
	if err := c.Run(ctx, chromedp.Tasks{
		chromedp.Sleep(2 * time.Second),
		chromedp.WaitVisible(".btn-success", chromedp.ByQuery),
		chromedp.AttributeValue(".btn-success", "href", downloadURL, &downloadOK, chromedp.ByQuery),
	}); err != nil {
		return fmt.Errorf("error getting song download URL: %v", err)
	}

	return nil
}
