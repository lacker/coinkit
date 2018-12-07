let http = require("http");

let WebTorrent = require("webtorrent-hybrid");

let proxy = http.createServer((req, res) => {
  res.end("this is the black hole");
});
proxy.listen(3333);

// Seed a WebTorrent
let buf = new Buffer("Hello decentralized world!");
buf.name = "hello.txt";
let client = new WebTorrent();
client.seed(buf, torrent => {
  console.log("info hash: " + torrent.infoHash);
  console.log("magnet uri: " + torrent.magnetURI);
});
