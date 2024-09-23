#!/bin/bash

# Fetch JSON data from the URL
json_data=$(curl -s "https://www2.vavoo.to/live2/index?countries=all&output=json")

# Create or clear the index.m3u file
echo "#EXTM3U" > index.m3u

# Create an associative array to keep track of groups
declare -A groups

# Parse JSON data and generate M3U files
# Function to process a single item
process_item() {
    local item="$1"
    local group=$(echo "$item" | jq -r '.group')
    local name=$(echo "$item" | jq -r '.name')
    local url=$(echo "$item" | jq -r '.url')
    local logo=$(echo "$item" | jq -r '.logo')
    local tvg_id=$(echo "$item" | jq -r '.tvg_id')

    # Display each channel obtained
    echo "Processing channel: $name"

    # Create or append to the group-specific M3U file
    if [ ! -f "index_${group}.m3u" ]; then
        echo "#EXTM3U" > "index_${group}.m3u"
    fi

    # Common M3U entry for all players
    {
        echo "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-name=\"$name\" tvg-logo=\"$logo\" group-title=\"$group\",$name"
        echo "#KODIPROP:inputstream=inputstream.adaptive"
        echo "#KODIPROP:inputstream.adaptive.manifest_type=hls"
        echo "#KODIPROP:http-user-agent=VAVOO/1.0"
        echo "#KODIPROP:http-referrer=https://vavoo.to/"
        echo "#EXTVLCOPT:http-user-agent=VAVOO/1.0"
        echo "#EXTVLCOPT:http-referrer=https://vavoo.to/"
        echo "#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://vavoo.to/\"}"
        echo "${url}|User-Agent=VAVOO/1.0&Referer=https://vavoo.to/"
    } >> "index_${group}.m3u"

    # Append to the main index.m3u file
    {
        echo "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-name=\"$name\" tvg-logo=\"$logo\" group-title=\"$group\",$name"
        echo "#KODIPROP:inputstream=inputstream.adaptive"
        echo "#KODIPROP:inputstream.adaptive.manifest_type=hls"
        echo "#KODIPROP:http-user-agent=VAVOO/1.0"
        echo "#KODIPROP:http-referrer=https://vavoo.to/"
        echo "#EXTVLCOPT:http-user-agent=VAVOO/1.0"
        echo "#EXTVLCOPT:http-referrer=https://vavoo.to/"
        echo "#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://vavoo.to/\"}"
        echo "${url}|User-Agent=VAVOO/1.0&Referer=https://vavoo.to/"
    } >> index.m3u

    # Track the group
    echo "$group" >> groups.txt
}

export -f process_item

# Fetch JSON data from the URL
json_data=$(curl -s "https://www2.vavoo.to/live2/index?countries=all&output=json")

# Create or clear the index.m3u file
echo "#EXTM3U" > index.m3u

# Create a temporary file for groups
> groups.txt

# Parse JSON data and generate M3U files using parallel
echo "$json_data" | jq -c '.[]' | parallel -j 512 process_item

# Create unique list of groups
sort -u groups.txt | while read group; do
    groups["$group"]=1
done

# Clean up temporary file
rm groups.txt

# Generate the HTML index file
echo "<html><body><h1>Index of M3U Files</h1><ul>" > index.html
echo "<li><a href=\"index.m3u\">index.m3u</a></li>" >> index.html
for group in "${!groups[@]}"; do
    echo "<li><a href=\"index_${group}.m3u\">index_${group}.m3u</a></li>" >> index.html
done
echo "</ul></body></html>" >> index.html

