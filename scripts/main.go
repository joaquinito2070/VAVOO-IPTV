package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
    "github.com/tebeka/selenium"
    "time"
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
func generateM3U(group, name, logo, tvgID, url string) (string, string) {
    url = strings.Replace(url, ".ts", "/index.m3u8", -1)
    url = strings.Replace(url, "/live2/play", "/play", -1)

    if strings.Contains(url, ".ts") {
        url = strings.Replace(url, ".ts", "/index.m3u8", -1)
    }
    if !strings.HasSuffix(url, "/index.m3u8") {
        url = url + "/index.m3u8"
    }

    url = strings.Replace(url, ".m3u8.m3u8", ".m3u8", -1)
    url = strings.Replace(url, "https://vavoo.to/play/", "https://joaquinito02.es/vavoo/", 1)
    url = strings.Replace(url, "/index.m3u8", ".m3u8", 1)

    return fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"%s\" http-user-agent=\"VAVOO/1.0\" http-referrer=\"https://vavoo.to/\",%s\n"+
        "#EXTVLCOPT:http-user-agent=VAVOO/1.0\n"+
        "#EXTVLCOPT:http-referrer=https://vavoo.to/\n"+
        "#KODIPROP:http-user-agent=VAVOO/1.0\n"+
        "#KODIPROP:http-referrer=https://vavoo.to/\n"+
        "#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://vavoo.to/\"}\n"+
        "%s", tvgID, name, logo, group, name, url), url
}

// fetchJSONData fetches JSON data using ChromeDriver
func fetchJSONData() ([]byte, error) {
    service, err := selenium.NewChromeDriverService("/usr/local/bin/chromedriver", 4444)
    if err != nil {
        return nil, fmt.Errorf("Error starting ChromeDriver service: %v", err)
    }
    defer service.Stop()

    caps := selenium.Capabilities{
        "browserName": "chrome",
        "goog:chromeOptions": map[string]interface{}{
            "args": []string{
                "--headless=new",
                "--no-sandbox",
                "--disable-dev-shm-usage",
                "--disable-gpu",
                "--remote-debugging-port=9222",
                "--disable-extensions",
            },
            "binary": "/usr/bin/google-chrome",
        },
    }

    driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
    if err != nil {
        return nil, fmt.Errorf("Error creating WebDriver: %v", err)
    }
    defer driver.Quit()

    fmt.Println("Starting download...")
    err = driver.Get("https://www2.vavoo.to/live2/index?countries=all&output=json")
    if err != nil {
        return nil, fmt.Errorf("Error navigating to URL: %v", err)
    }

    time.Sleep(5 * time.Second)

    // Get the pre element that contains the JSON
    element, err := driver.FindElement(selenium.ByTagName, "pre")
    if err != nil {
        return nil, fmt.Errorf("Error finding JSON element: %v", err)
    }

    // Get the text content of the pre element
    jsonContent, err := element.Text()
    if err != nil {
        return nil, fmt.Errorf("Error getting JSON content: %v", err)
    }

    return []byte(jsonContent), nil
}

// processItem processes a single item and returns M3U content, group, and htaccess URL
func processItem(item Item) (string, string, string, error) {
    m3uContent, htaccessURL := generateM3U(item.Group, item.Name, item.Logo, item.TvgID, item.URL)
    return m3uContent, item.Group, htaccessURL, nil
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

    if _, err := os.Stat("index.m3u"); err == nil {
        os.Remove("index.m3u")
    }

    indexM3U, err := os.Create("index.m3u")
    if err != nil {
        fmt.Printf("Error creating index.m3u: %v\n", err)
        return
    }
    defer indexM3U.Close()

    indexM3U.WriteString("#EXTM3U\n")
    groups := make(map[string]*os.File)
    processedCount := 0
    var idsContent string

    for _, item := range items {
        m3uContent, group, _, err := processItem(item)
        if err != nil {
            fmt.Printf("Error processing item: %v\n", err)
            continue
        }

        if _, exists := groups[group]; !exists {
            groupFileName := "index_" + group + ".m3u"
            groupFile, err := os.Create(groupFileName)
            if err != nil {
                fmt.Printf("Error creating group file: %v\n", err)
                continue
            }
            groupFile.WriteString("#EXTM3U\n")
            groups[group] = groupFile
        }

        indexM3U.WriteString(m3uContent + "\n")
        groups[group].WriteString(m3uContent + "\n")

        id := strings.TrimPrefix(item.URL, "https://vavoo.to/live2/play/")
        id = strings.TrimSuffix(id, ".ts")
        idsContent += id + "\n"

        processedCount++
        fmt.Printf("Processed %d/%d channels\n", processedCount, len(items))
    }

    for _, groupFile := range groups {
        groupFile.Close()
    }

    err = ioutil.WriteFile("ids.txt", []byte(idsContent), 0644)
    if err != nil {
        fmt.Printf("Error writing ids.txt: %v\n", err)
        return
    }

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

    fmt.Println("M3U files, ids.txt, and HTML index generated successfully.")
}
