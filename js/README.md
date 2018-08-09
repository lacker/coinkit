# Installing the chrome extension locally

1. Make sure you have the local servers running: `./start-local.sh`

2. Build the extension locally
   a. `npm install`
   b. `npm run build`

3. Load the extension in Chrome
   a. Navigate to `chrome://extensions`
   b. Toggle "Developer Mode" on, if it's not on already
   c. Click "Load Unpackaged" and select the `extension` directory