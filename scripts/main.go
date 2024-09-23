package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type Channel struct {
	Group string `json:"group"`
	Logo  string `json:"logo"`
	Name  string `json:"name"`
	TvgID string `json:"tvg_id"`
	URL   string `json:"url"`
}

func main() {
	// Fetch JSON data
	resp, err := http.Get("https://www2.vavoo.to/live2/index?countries=all&output=json")
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	var channels []Channel
	err = json.Unmarshal(body, &channels)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Create maps to store group-specific and all channels
	groupChannels := make(map[string][]Channel)
	allChannels := make([]Channel, 0)

	for _, channel := range channels {
		groupChannels[channel.Group] = append(groupChannels[channel.Group], channel)
		allChannels = append(allChannels, channel)
	}

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup

	// Create a channel to limit concurrency
	semaphore := make(chan struct{}, 512)

	// Create progress bar
	bar := progressbar.Default(int64(len(groupChannels) + 1))

	// Generate M3U files for each group
	for group, channels := range groupChannels {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(group string, channels []Channel) {
			defer wg.Done()
			defer func() { <-semaphore }()
			generateM3U(fmt.Sprintf("index_%s.m3u", group), channels)
			bar.Add(1)
		}(group, channels)
	}

	// Generate M3U file for all channels
	wg.Add(1)
	semaphore <- struct{}{}
	go func() {
		defer wg.Done()
		defer func() { <-semaphore }()
		generateM3U("index.m3u", allChannels)
		bar.Add(1)
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	// Generate HTML file
	generateHTML()

	fmt.Println("M3U and HTML files generated successfully.")
}

func generateM3U(filename string, channels []Channel) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	file.WriteString("#EXTM3U\n")

	for _, channel := range channels {
		url := strings.Replace(channel.URL, "/live2/play", "/play", 1)
		url = strings.Replace(url, ".ts", "/index.m3u8", 1)

		file.WriteString(fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-logo=\"%s\" group-title=\"%s\",Referer=\"https://vavoo.tv\" User-Agent=\"VAVOO/1.0\",%s\n", channel.TvgID, channel.Logo, channel.Group, channel.Name))
		file.WriteString(fmt.Sprintf("%s\n", url))
	}
}

func generateHTML() {
	files, err := filepath.Glob("index*.m3u")
	if err != nil {
		fmt.Println("Error finding M3U files:", err)
		return
	}

	htmlContent := "<html><body><h1>M3U Playlists</h1><ul>"
	for _, file := range files {
		htmlContent += fmt.Sprintf("<li><a href='%s'>%s</a></li>", file, file)
	}
	htmlContent += "</ul></body></html>"

	err = ioutil.WriteFile("index.html", []byte(htmlContent), 0644)
	if err != nil {
		fmt.Println("Error writing HTML file:", err)
	}
}
