// An integration test
const WebTorrent = require("webtorrent-hybrid");
const TorrentClient = require("./TorrentClient.js");

// TODO: figure out why this doesn't shut down cleanly

const SAMPLESITE =
  "magnet:?xt=urn:btih:e60f82343019bd711c5c731b46e118b0f2b2ecc6&dn=samplesite&tr=ws%3A%2F%2Flocalhost%3A4444&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com";

let client = new WebTorrent();
let info = SAMPLESITE;
console.log("adding", info);
client.add(info, torrent => {
  console.log("torrent metadata is ready");
  for (let file of torrent.files) {
    console.log("file:", file.name);
  }

  torrent.on("done", () => {
    console.log("downloaded", torrent.downloaded, "bytes");
    console.log("shutting down");
    torrent.destroy();
  });
});
