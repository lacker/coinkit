const WebTorrent = require("webtorrent-hybrid");

// Run a black hole proxy - this just returns a blank document for every URL.
// This is needed so that the extension can pretend a URL that doesn't resolve is
// actually a functioning web page.

// TODO: see if we can just muck with the error page

// Seed a WebTorrent
let buf = new Buffer("Hello decentralized world!");
buf.name = "hello.txt";
let client = new WebTorrent();
client.seed(buf, torrent => {
  console.log("info hash: " + torrent.infoHash);
  console.log("magnet uri: " + torrent.magnetURI);
});
