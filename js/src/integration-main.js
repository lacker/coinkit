// An integration test
const WebTorrent = require("webtorrent");
const TorrentClient = require("./TorrentClient.js");

const SINTEL =
  "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent";

async function main() {
  let client = new TorrentClient();
  let magnet = SINTEL;
  let torrent = client.download(magnet);
  await torrent.monitorProgress();
  client.destroy();
}

main()
  .then(() => {
    console.log("done");
  })
  .catch(e => {
    console.log("Unhandled " + e);
    process.exit(1);
  });
