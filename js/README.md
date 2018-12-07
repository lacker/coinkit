# Overview

There are several different build targets produced by this JavaScript codebase.

The chrome extension is built into `/extension/`.

The sample app is built into `/app/`.

# Installing the chrome extension locally

1. Make sure you have the local servers running: `./start-local.sh`

2. Build the extension locally
   a. `npm install`
   b. `npm run extension` (which will watch for changes and build continuously)

3. Load the extension in Chrome
   a. Navigate to `chrome://extensions`
   b. Toggle "Developer Mode" on, if it's not on already
   c. Click "Load Unpacked" and select the `coinkit/js/extension` directory

4. Run the sample app to try things out
   a. `npm run app`
   b. Go to `localhost:1234`