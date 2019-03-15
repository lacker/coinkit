// This is the entry point for the hosting server.

// TODO: make this read args and create a HostingServer

const http = require("http");
const path = require("path");
const WebTorrent = require("webtorrent-hybrid");

const BlackHoleProxy = require("./BlackHoleProxy.js");

/* TODO: remove this webtorrent seeding once the deploy-based seeding is working

// Seed a WebTorrent
let client = new WebTorrent();
let dir = path.resolve(__dirname, "samplesite");
client.seed(dir, torrent => {
  console.log("info hash: " + torrent.infoHash);
  console.log("magnet uri: " + torrent.magnetURI);

  // fakeChain tells clients where to look for the torrent
  // TODO: make this happen via the real chain
  let fakeChainPort = 4444;
  let fakeChain = http.createServer((req, res) => {
    res.end(
      JSON.stringify({
        magnet: torrent.magnetURI
      })
    );
  });
  fakeChain.listen(fakeChainPort);
  console.log("running fake chain on port", fakeChainPort);
});

*/

let proxy = new BlackHoleProxy(3333);
