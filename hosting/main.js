const WebTorrent = require("webtorrent-hybrid");

let buf = new Buffer("Hello decentralized world!");
buf.name = "hello.txt";
let client = new WebTorrent();
client.seed(buf, () => {
  console.log("seeding has begun");
});
