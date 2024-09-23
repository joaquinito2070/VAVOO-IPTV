package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Item represents a single channel item
type Item struct {
	Group string `json:"group"`
	Name  string `json:"name"`
	Logo  string `json:"logo"`
	TvgID string `json:"tvg_id"`
	URL   string `json:"url"`
}

// generateM3U generates M3U content for a single item
func generateM3U(group, name, logo, tvgID, url string) string {
	// Replace .ts with /index.m3u8 and /live2/play with /play
	url = strings.Replace(url, ".ts", "/index.m3u8", -1)
	url = strings.Replace(url, "/live2/play", "/play", -1)

	return fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"%s\" http-user-agent=\"VAVOO/1.0\" http-referrer=\"https://vavoo.to/\",%s\n"+
		"#EXTVLCOPT:http-user-agent=VAVOO/1.0\n"+
		"#EXTVLCOPT:http-referrer=https://vavoo.to/\n"+
		"#KODIPROP:http-user-agent=VAVOO/1.0\n"+
		"#KODIPROP:http-referrer=https://vavoo.to/\n"+
		"#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://vavoo.to/\"}\n"+
		"%s", tvgID, name, logo, group, name, url)
}

// fetchJSONData fetches JSON data from the specified URL
func fetchJSONData() ([]byte, error) {
	resp, err := http.Get("https://www2.vavoo.to/live2/index?countries=all&output=json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// processItem processes a single item and returns M3U content and group
func processItem(item Item) (string, string, error) {
	m3uContent := generateM3U(item.Group, item.Name, item.Logo, item.TvgID, item.URL)
	return m3uContent, item.Group, nil
}

func main() {
	jsonData, err := fetchJSONData()
	if err != nil {
		fmt.Printf("Error fetching JSON data: %v\n", err)
		return
	}

	var items []Item
	err = json.Unmarshal(jsonData, &items)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return
	}

	indexM3U, err := os.Create("index.m3u")
	if err != nil {
		fmt.Printf("Error creating index.m3u: %v\n", err)
		return
	}
	defer indexM3U.Close()

	indexM3U.WriteString("#EXTM3U\n")

	groups := make(map[string]bool)
	processedCount := 0

	for _, item := range items {
		m3uContent, group, err := processItem(item)
		if err != nil {
			fmt.Printf("Error processing item: %v\n", err)
			continue
		}

		groups[group] = true

		indexM3U.WriteString(m3uContent + "\n")

		groupFile, err := os.OpenFile("index_"+group+".m3u", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening group file: %v\n", err)
			continue
		}
		groupFile.WriteString(m3uContent + "\n")
		groupFile.Close()

		processedCount++
		fmt.Printf("Processed %d/%d channels\n", processedCount, len(items))
	}

	// Generate HTML
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>M3U Playlists</title>
</head>
<body>
    <h1>M3U Playlists</h1>
    <h2><a href="index.m3u">Complete Playlist</a></h2>
    <h2>Group-specific Playlists:</h2>
    <ul>
`

	for group := range groups {
		html += fmt.Sprintf("        <li><a href=\"index_%s.m3u\">%s</a></li>\n", group, group)
	}

	html += `    </ul>
</body>
</html>`

	err = ioutil.WriteFile("index.html", []byte(html), 0644)
	if err != nil {
		fmt.Printf("Error writing index.html: %v\n", err)
		return
	}

	fmt.Println("M3U files and HTML index generated successfully.")
}
