const fs = require('fs');
const https = require('https');
const { promisify } = require('util');
const { Worker, isMainThread, parentPort, workerData } = require('worker_threads');

const writeFileAsync = promisify(fs.writeFile);
const appendFileAsync = promisify(fs.appendFile);

// Function to generate M3U content
function generateM3U(group, name, logo, tvgId, url) {
    return `#EXTINF:-1 tvg-id="${tvgId}" tvg-name="${name}" tvg-logo="${logo}" group-title="${group}" http-user-agent="VAVOO/1.0" http-referrer="https://www.vavoo.to/",${name}
#EXTVLCOPT:http-user-agent=VAVOO/1.0
#EXTVLCOPT:http-referrer=https://www.vavoo.to/
#KODIPROP:http-user-agent=VAVOO/1.0
#KODIPROP:http-referrer=https://www.vavoo.to/
#EXTHTTP:{"User-Agent":"VAVOO/1.0","Referer":"https://www.vavoo.to/"}
${url}`;
}

// Fetch JSON data
function fetchJSONData() {
    return new Promise((resolve, reject) => {
        https.get('https://www2.vavoo.to/live2/index?countries=all&output=json', (res) => {
            let data = '';
            res.on('data', (chunk) => data += chunk);
            res.on('end', () => resolve(data));
        }).on('error', reject);
    });
}

// Process a single item
function processItem(item) {
    try {
        const { group = '', name = '', logo = '', tvg_id = '', url = '' } = JSON.parse(item);
        const m3uContent = generateM3U(group, name, logo, tvg_id, url);
        return { m3uContent, group };
    } catch (error) {
        console.error('Invalid JSON:', item);
        return null;
    }
}

// Worker thread function
if (!isMainThread) {
    parentPort.on('message', (item) => {
        const result = processItem(item);
        parentPort.postMessage(result);
    });
}

async function main() {
    try {
        const jsonData = await fetchJSONData();
        const items = JSON.parse(jsonData);

        await writeFileAsync('index.m3u', '#EXTM3U\n');
        await writeFileAsync('groups.txt', '');

        const numWorkers = 512;
        const workers = new Array(numWorkers).fill().map(() => new Worker(__filename));

        const groups = new Set();
        let processedCount = 0;

        const processChunk = (chunk) => {
            return new Promise((resolve) => {
                const results = [];
                let completed = 0;

                chunk.forEach((item, index) => {
                    workers[index].postMessage(JSON.stringify(item));
                    workers[index].once('message', (result) => {
                        if (result) {
                            results.push(result);
                            groups.add(result.group);
                        }
                        completed++;
                        if (completed === chunk.length) {
                            resolve(results);
                        }
                    });
                });
            });
        };

        for (let i = 0; i < items.length; i += numWorkers) {
            const chunk = items.slice(i, i + numWorkers);
            const results = await processChunk(chunk);

            for (const { m3uContent, group } of results) {
                await appendFileAsync('index.m3u', m3uContent + '\n');
                await appendFileAsync(`index_${group}.m3u`, m3uContent + '\n');
            }

            processedCount += chunk.length;
            console.log(`Processed ${processedCount}/${items.length} channels`);
        }

        workers.forEach(worker => worker.terminate());

        // Generate HTML
        let html = `<!DOCTYPE html>
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
`;

        for (const group of groups) {
            html += `        <li><a href="index_${group}.m3u">${group}</a></li>\n`;
        }

        html += `    </ul>
</body>
</html>`;

        await writeFileAsync('index.html', html);

        console.log("M3U files and HTML index generated successfully.");
    } catch (error) {
        console.error("An error occurred:", error);
    }
}

if (isMainThread) {
    main();
}
