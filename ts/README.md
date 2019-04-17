# About This Code

This directory contains the TypeScript codebase. There are several different build targets
produced by the same codebase. Some targets build for a browser
environment; some build for a node environment.

`src/browser/` contains code that is only intended to run in the
browser.

`src/node/` contains code that is only intended to run in node.js.

`src/iso/` contains code that is intended to run in either node.js or
the browser.

One key exception to this rule is that unit tests are always running
in node.js, albeit from a Jest environment that is mimicking the
browser.

TypeScript will store its build files in `/build/`, and Parcel will
store its build files in `/dist/`. You can ignore these directories
unless you are mucking around with the build system.

# Build Targets

The local chrome extension, used for testing against a blockchain
running on your machine, is built into `/ext-local/` with the command
`npm run ext-local`.

The CLI is run as a Node application via `npm run cli`.

The hosting server is run as a Node application via `npm run hserver`.

There is an app used for testing things, built into `/app/` with the
command `npm run app`. You don't really need this; it's just handy to
use as a testbed sometimes.

# Testing Things Locally

1. Make sure you have the local blockchain running:
`./start-local.sh`. If this isn't working, see the README one
directory up.

2. Build the extension locally
   a. `npm install`
   b. `npm run ext-local`

3. Load the extension in Chrome
   a. Navigate to `chrome://extensions`
   b. Toggle "Developer Mode" on
   c. Click "Load Unpacked" and select the `coinkit/ts/ext-local` directory

4. Run a hosting server
   a. `npm run hserver`

5. Deploy the "hello" website. Use the passphrase "mint"
   a. `npm run cli create-bucket`
   b. `npm run cli deploy ./samplesite hello

6. Test that it works
   a. Navigate to "hello.coinkit" in your browser
   b. You should see a website with a picture of a frog