# GoScreenShot

Geguti is a simple HTTP-based service written in Go that captures screenshots (currently in development :D ) of webpages using [chromedp](https://github.com/chromedp/chromedp). 
It provides an API endpoint to take screenshots of any given URL and saves them locally.

## Features

- Capture full-page screenshots of webpages.
- Configure output directory with an environment variable.


## How It Works

1. Send a POST request to `/screenshot` with a JSON payload:
   ```json
   {
       "url": "https://example.com",
       "timeout": 30
   }


