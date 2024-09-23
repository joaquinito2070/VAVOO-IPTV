<?php

// Function to generate M3U content
function generate_m3u($group, $name, $logo, $tvg_id, $url) {
    // Modify URL to replace /live2/play/ with /play/ and .ts with /index.m3u8
    $url = str_replace('https://vavoo.to/live2/play/', 'https://vavoo.to/play/', $url);
    $url = str_replace('.ts', '/index.m3u8', $url);

    return "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-name=\"$name\" tvg-logo=\"$logo\" group-title=\"$group\" http-user-agent=\"VAVOO/1.0\" http-referrer=\"https://vavoo.to/\",$name
#EXTVLCOPT:http-user-agent=VAVOO/1.0
#EXTVLCOPT:http-referrer=https://vavoo.to/
#KODIPROP:http-user-agent=VAVOO/1.0
#KODIPROP:http-referrer=https://vavoo.to/
#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://vavoo.to/\"}
$url";
}

// Fetch JSON data
function fetch_json_data() {
    $json_data = file_get_contents('https://www2.vavoo.to/live2/index?countries=all&output=json');
    return $json_data;
}

// Process a single item
function process_item($item) {
    try {
        $group = $item['group'] ?? '';
        $name = $item['name'] ?? '';
        $logo = $item['logo'] ?? '';
        $tvg_id = $item['tvg_id'] ?? '';
        $url = $item['url'] ?? '';
        // Modify URL to replace /live2/play/ with /play/ and .ts with /index.m3u8
        $modified_url = str_replace('https://vavoo.to/live2/play/', 'https://vavoo.to/play/', $url);
        $modified_url = str_replace('.ts', '/index.m3u8', $modified_url);
        $m3u_content = generate_m3u($group, $name, $logo, $tvg_id, $modified_url);
        return [$m3u_content, $group];
    } catch (Exception $e) {
        echo "Error processing item: " . $e->getMessage() . "\n";
        return null;
    }
}

function main() {
    try {
        $json_data = fetch_json_data();
        $items = json_decode($json_data, true);

        file_put_contents('index.m3u', "#EXTM3U\n");

        $groups = [];
        $processed_count = 0;

        foreach ($items as $item) {
            $result = process_item($item);
            if ($result) {
                list($m3u_content, $group) = $result;
                $groups[$group] = true;
                file_put_contents('index.m3u', $m3u_content . "\n", FILE_APPEND);
                file_put_contents("index_{$group}.m3u", $m3u_content . "\n", FILE_APPEND);
                $processed_count++;
                echo "Processed $processed_count/" . count($items) . " channels\n";
            }
        }

        // Generate HTML
        $html = "<!DOCTYPE html>
<html lang=\"en\">
<head>
    <meta charset=\"UTF-8\">
    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">
    <title>M3U Playlists</title>
</head>
<body>
    <h1>M3U Playlists</h1>
    <h2><a href=\"index.m3u\">Complete Playlist</a></h2>
    <h2>Group-specific Playlists:</h2>
    <ul>
";

        foreach (array_keys($groups) as $group) {
            $html .= "        <li><a href=\"index_{$group}.m3u\">{$group}</a></li>\n";
        }

        $html .= "    </ul>
</body>
</html>";

        file_put_contents('index.html', $html);

        echo "M3U files and HTML index generated successfully.\n";
    } catch (Exception $e) {
        echo "An error occurred: " . $e->getMessage() . "\n";
    }
}

main();
