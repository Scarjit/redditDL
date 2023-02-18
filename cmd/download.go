package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var processChan = make(chan Post, 100)

var downloadedPostHashes []string
var finishedInserting bool

func Downloader() {
	fmt.Println("Starting downloader...")
	var posts []Post
	DB.Find(&posts)
	loadDownloadedImages()

	go startDownloaders()
	for _, post := range posts {
		processChan <- post
	}
	finishedInserting = true
}

func loadDownloadedImages() {
	fmt.Println("Loading downloaded images...")
	files, err := os.ReadDir("./images")
	if err != nil {
		log.Fatal(err)
	}
	// pre-allocation
	downloadedPostHashes = make([]string, 0, len(files))

	for _, file := range files {
		downloadedPostHashes = append(downloadedPostHashes, file.Name())
	}
	fmt.Println("Loaded downloaded images")
}

func startDownloaders() {
	fmt.Println("Starting downloaders...")
	var waitGroup sync.WaitGroup
	waitGroup.Add(16)
	for i := 0; i < 16; i++ {
		go downloader(&waitGroup)
	}
	fmt.Println("Downloaders started")
	waitGroup.Wait()
	fmt.Println("Downloaders finished")
}

func downloader(wg *sync.WaitGroup) {
	fmt.Println("Downloader started")
	time.Sleep(1 * time.Second)
	// Select from channel if not empty
	for {
		select {
		case post := <-processChan:
			// Check if image has already been downloaded
			if !contains(downloadedPostHashes, post.FullId) {
				// Download image
				DownloadImage(post.Url, post.FullId)
				// Add to downloadedPostHashes
				downloadedPostHashes = append(downloadedPostHashes, post.FullId)
			}
		default:
			if finishedInserting {
				// If channel is empty, exit
				wg.Done()
				return
			} else {
				fmt.Printf("Downloader waiting...\n")
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func contains(haystack []string, id string) bool {
	for _, s := range haystack {
		if s == id {
			return true
		}
	}
	return false
}

func DownloadImage(url string, id string) {
	// Check if url is valid
	if !isValidUrl(url) {
		return
	}
	// Check if url is an image
	if !isImage(url) {
		return
	}

	// Download image
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading image: %s", err)
	}
	defer resp.Body.Close()

	// Create file
	var file *os.File
	file, err = os.Create(fmt.Sprintf("./images/%s.jpg", id))
	if err != nil {
		fmt.Printf("Error creating file: %s", err)
	}
	defer file.Close()

	// Read response body
	var bytes []byte
	bytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	// Write to file
	_, err = file.Write(bytes)
	if err != nil {
		fmt.Printf("Error writing to file: %s", err)
	}
	fmt.Printf("Downloaded image: %s\n", id)
}

func isImage(u string) bool {
	// Do a HEAD request to check if the url is an image
	head, err := http.Head(u)
	if err != nil {
		return false
	}
	get := head.Header.Get("Content-Type")
	return strings.HasPrefix(get, "image/")
}

func isValidUrl(u string) bool {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	return true
}
