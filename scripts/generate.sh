#!/bin/bash

# Function to generate M3U content
generate_m3u() {
    local group="$1"
    local name="$2"
    local logo="$3"
    local tvg_id="$4"
    local url="$5"
    
    echo "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-name=\"$name\" tvg-logo=\"$logo\" group-title=\"$group\" http-user-agent=\"VAVOO/1.0\" http-referrer=\"https://www.vavoo.to/\",$name"
    echo "#EXTVLCOPT:http-user-agent=VAVOO/1.0"
    echo "#EXTVLCOPT:http-referrer=https://www.vavoo.to/"
    echo "#KODIPROP:http-user-agent=VAVOO/1.0"
    echo "#KODIPROP:http-referrer=https://www.vavoo.to/"
    echo "#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://www.vavoo.to/\"}"
    echo "$url"
}

# Fetch JSON data
json_data=$(curl -s "https://www2.vavoo.to/live2/index?countries=all&output=json")

# Generate index.m3u and group-specific M3U files
echo "#EXTM3U" > index.m3u
echo "" > groups.txt

echo "$json_data" | jq -c '.[]' | parallel -j 1024 '
    group=$(echo {} | jq -r .group)
    name=$(echo {} | jq -r .name)
    logo=$(echo {} | jq -r .logo)
    tvg_id=$(echo {} | jq -r .tvg_id)
    url=$(echo {} | jq -r .url)

    m3u_content=$(echo "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-name=\"$name\" tvg-logo=\"$logo\" group-title=\"$group\" http-user-agent=\"VAVOO/1.0\" http-referrer=\"https://www.vavoo.to/\",$name"; \
                  echo "#EXTVLCOPT:http-user-agent=VAVOO/1.0"; \
                  echo "#EXTVLCOPT:http-referrer=https://www.vavoo.to/"; \
                  echo "#KODIPROP:http-user-agent=VAVOO/1.0"; \
                  echo "#KODIPROP:http-referrer=https://www.vavoo.to/"; \
                  echo "#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://www.vavoo.to/\"}"; \
                  echo "$url")
    
    echo "$m3u_content" >> index.m3u
    echo "$m3u_content" >> "index_${group}.m3u"
    echo "$group" >> groups.txt
'

# Remove duplicate groups
sort -u groups.txt > unique_groups.txt

# Generate HTML
cat << EOF > index.html
<!DOCTYPE html>
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
EOF

while read -r group; do
    echo "        <li><a href=\"index_${group}.m3u\">$group</a></li>" >> index.html
done < unique_groups.txt

cat << EOF >> index.html
    </ul>
</body>
</html>
EOF

# Clean up temporary files
rm groups.txt unique_groups.txt

echo "M3U files and HTML index generated successfully."
