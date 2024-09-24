<?php
// Read the ids.txt file
$ids = file('ids.txt', FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);

// Get the requested ID from the URL
$requestUri = $_SERVER['REQUEST_URI'];
$matches = [];
preg_match('/\/vavoo\/(.+)\.m3u8/', $requestUri, $matches);
$id = $matches[1] ?? null;

if ($id && in_array($id, $ids)) {
    // Set CORS headers
    header('Access-Control-Allow-Origin: *');
    header('Access-Control-Allow-Methods: GET, OPTIONS');
    header('Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization, X-Request-ID, X-Joaquinito02-Trace');

    // Set custom headers
    header('X-Request-ID: ' . uniqid());
    header('X-Joaquinito02-Trace: ' . bin2hex(random_bytes(16)));

    // Redirect to the new URL
    $newUrl = "https://vavoo.to/play/$id/index.m3u8";
    header("Location: $newUrl", true, 302);
    exit();
} else {
    // Return a 404 response if the ID is not found
    header("HTTP/1.1 404 Not Found");
    echo "ID not found.";
    exit();
}
?>
