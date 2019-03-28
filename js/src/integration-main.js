// An integration test
const WebTorrent = require("webtorrent-hybrid");
const TorrentClient = require("./TorrentClient.js");

const SAMPLESITE =
  "magnet:?xt=urn:btih:e60f82343019bd711c5c731b46e118b0f2b2ecc6&dn=samplesite&tr=ws%3A%2F%2Flocalhost%3A4444&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com";

async function main() {
  let client = new TorrentClient();
  let torrent = await client.download(SAMPLESITE);
  await torrent.monitorProgress();
  await client.destroy();
}

main().then(() => {
  console.log("done");
});
