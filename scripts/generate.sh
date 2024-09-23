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

    # Display each channel obtained
    echo "Processing channel: $name"

    # Create or append to the group-specific M3U file
    if [ ! -f "index_${group}.m3u" ]; then
        echo "#EXTM3U" > "index_${group}.m3u"
    fi

    # Common M3U entry for all players
    echo "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-name=\"$name\" tvg-logo=\"$logo\" group-title=\"$group\",$name" >> "index_${group}.m3u"
    
    # Kodi specific properties
    echo "#KODIPROP:inputstream=inputstream.adaptive" >> "index_${group}.m3u"
    echo "#KODIPROP:inputstream.adaptive.manifest_type=hls" >> "index_${group}.m3u"
    echo "#KODIPROP:http-user-agent=VAVOO/1.0" >> "index_${group}.m3u"
    echo "#KODIPROP:http-referrer=https://vavoo.to/" >> "index_${group}.m3u"
    
    # VLC specific options
    echo "#EXTVLCOPT:http-user-agent=VAVOO/1.0" >> "index_${group}.m3u"
    echo "#EXTVLCOPT:http-referrer=https://vavoo.to/" >> "index_${group}.m3u"
    
    # TVHeadend specific options
    echo "#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://vavoo.to/\"}" >> "index_${group}.m3u"
    
    # URL with User-Agent and Referer for players that support it in the URL
    echo "${url}|User-Agent=VAVOO/1.0&Referer=https://vavoo.to/" >> "index_${group}.m3u"

    # Append to the main index.m3u file
    echo "#EXTINF:-1 tvg-id=\"$tvg_id\" tvg-name=\"$name\" tvg-logo=\"$logo\" group-title=\"$group\",$name" >> index.m3u
    echo "#KODIPROP:inputstream=inputstream.adaptive" >> index.m3u
    echo "#KODIPROP:inputstream.adaptive.manifest_type=hls" >> index.m3u
    echo "#KODIPROP:http-user-agent=VAVOO/1.0" >> index.m3u
    echo "#KODIPROP:http-referrer=https://vavoo.to/" >> index.m3u
    echo "#EXTVLCOPT:http-user-agent=VAVOO/1.0" >> index.m3u
    echo "#EXTVLCOPT:http-referrer=https://vavoo.to/" >> index.m3u
    echo "#EXTHTTP:{\"User-Agent\":\"VAVOO/1.0\",\"Referer\":\"https://vavoo.to/\"}" >> index.m3u
    echo "${url}|User-Agent=VAVOO/1.0&Referer=https://vavoo.to/" >> index.m3u

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

