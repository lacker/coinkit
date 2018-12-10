let http = require("http");

let WebTorrent = require("webtorrent-hybrid");

// Seed a WebTorrent
let buf = new Buffer("Hello decentralized world!");
buf.name = "hello.txt";
let client = new WebTorrent();
client.seed(buf, torrent => {
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
  res.end(content);
});
proxy.listen(proxyPort);
console.log("running proxy on port", proxyPort);
