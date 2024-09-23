#!/bin/bash

# Fetch JSON data from the URL
json_data=$(curl -s "https://www2.vavoo.to/live2/index?countries=all&output=json")

# Create or clear the index.m3u file
echo "#EXTM3U" > index.m3u

# Create an associative array to keep track of groups
declare -A groups

# Parse JSON data and generate M3U files
echo "$json_data" | jq -c '.[]' | while read -r item; do
    group=$(echo "$item" | jq -r '.group')
    name=$(echo "$item" | jq -r '.name')
    url=$(echo "$item" | jq -r '.url')
    logo=$(echo "$item" | jq -r '.logo')
    tvg_id=$(echo "$item" | jq -r '.tvg_id')

    # Create or append to the group-specific M3U file
    if [ ! -f "index_${group}.m3u" ]; then
        echo "#EXTM3U" > "index_${group}.m3u"
    fi
    echo "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-logo=\"$logo\" group-title=\"$group\", $name" >> "index_${group}.m3u"
    echo "$url" >> "index_${group}.m3u"

    # Append to the main index.m3u file
    echo "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-logo=\"$logo\" group-title=\"$group\", $name" >> index.m3u
    echo "$url" >> index.m3u

    # Track the group
    groups["$group"]=1
done

# Generate the HTML index file
echo "<html><body><h1>Index of M3U Files</h1><ul>" > index.html
echo "<li><a href=\"index.m3u\">index.m3u</a></li>" >> index.html
for group in "${!groups[@]}"; do
    echo "<li><a href=\"index_${group}.m3u\">index_${group}.m3u</a></li>" >> index.html
done
echo "</ul></body></html>" >> index.html

