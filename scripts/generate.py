import asyncio
import aiohttp
import aiofiles
from concurrent.futures import ThreadPoolExecutor

# Function to generate M3U content
def generate_m3u(group, name, logo, tvg_id, url):
    # Modify URL to replace /live2/play/ with /play/ and .ts with /index.m3u8
    url = url.replace('https://vavoo.to/live2/play/', 'https://vavoo.to/play/')
    url = url.replace('.ts', '/index.m3u8')

    return f"""#EXTINF:-1 tvg-id="{tvg_id}" tvg-name="{name}" tvg-logo="{logo}" group-title="{group}" http-user-agent="VAVOO/1.0" http-referrer="https://vavoo.to/",{name}
#EXTVLCOPT:http-user-agent=VAVOO/1.0
#EXTVLCOPT:http-referrer=https://vavoo.to/
#KODIPROP:http-user-agent=VAVOO/1.0
#KODIPROP:http-referrer=https://vavoo.to/
#EXTHTTP:{{"User-Agent":"VAVOO/1.0","Referer":"https://vavoo.to/"}}
{url}"""

# Fetch JSON data
async def fetch_json_data(session):
    async with session.get('https://www2.vavoo.to/live2/index?countries=all&output=json') as response:
        return await response.text()

# Process a single item
def process_item(item):
    try:
        group = item.get('group', '')
        name = item.get('name', '')
        logo = item.get('logo', '')
        tvg_id = item.get('tvg_id', '')
        url = item.get('url', '')
        modified_url = url.replace('.ts', '/index.m3u8').replace('/live2/play', '/play')
        m3u_content = generate_m3u(group, name, logo, tvg_id, modified_url)
        return m3u_content, group
    except Exception as e:
        print(f'Error processing item: {e}')
        return None

async def main():
    async with aiohttp.ClientSession() as session:
        try:
            json_data = await fetch_json_data(session)
            items = json.loads(json_data)

            async with aiofiles.open('index.m3u', 'w') as f:
                await f.write('#EXTM3U\n')

            groups = set()
            processed_count = 0

            with ThreadPoolExecutor(max_workers=512) as executor:
                loop = asyncio.get_event_loop()
                tasks = [loop.run_in_executor(executor, process_item, item) for item in items]
                results = await asyncio.gather(*tasks)

            for result in results:
                if result:
                    m3u_content, group = result
                    groups.add(group)
                    async with aiofiles.open('index.m3u', 'a') as f:
                        await f.write(m3u_content + '\n')
                    async with aiofiles.open(f'index_{group}.m3u', 'a') as f:
                        await f.write(m3u_content + '\n')
                    processed_count += 1
                    print(f'Processed {processed_count}/{len(items)} channels')

            # Generate HTML
            html = f"""<!DOCTYPE html>
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
{"".join(f'        <li><a href="index_{group}.m3u">{group}</a></li>\n' for group in groups)}
    </ul>
</body>
</html>"""

            async with aiofiles.open('index.html', 'w') as f:
                await f.write(html)

            print("M3U files and HTML index generated successfully.")
        except Exception as e:
            print(f"An error occurred: {e}")

if __name__ == "__main__":
    asyncio.run(main())
