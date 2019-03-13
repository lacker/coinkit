// This is the entry point for the hosting server.

const http = require("http");
const path = require("path");

const WebTorrent = require("webtorrent-hybrid");

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

// This code should never run because the document load gets canceled
let content = `
<html>
<head>
<script>
console.log("running black hole");
</script>
</head>
<body>
this is the black hole proxy
</body>
</html>
`;

let proxyPort = 3333;
let proxy = http.createServer((req, res) => {
  console.log("proxying", req.url);
  res.end(content);
});
proxy.listen(proxyPort);
console.log("running proxy on port", proxyPort);
