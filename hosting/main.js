const WebTorrent = require("webtorrent-hybrid");

let buf = new Buffer("Hello decentralized world!");
buf.name = "hello.txt";
let client = new WebTorrent();
client.seed(buf, torrent => {
  console.log("info hash: " + torrent.infoHash);
  console.log("magnet uri: " + torrent.magnetURI);
});
