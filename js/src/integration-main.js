// An integration test
const WebTorrent = require("webtorrent-hybrid");
const TorrentClient = require("./TorrentClient.js");

const SINTEL =
  "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent";

const DAT = "9ca28626647404a6fd831b56fb2e8530891a943e";

/* Works for sintel but that's it
async function main() {
  let client = new TorrentClient();
  client.verbose = true;
  let torrent = await client.download(SAMPLESITE);
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
*/

const SAMPLESITE = "e60f82343019bd711c5c731b46e118b0f2b2ecc6";

let client = new WebTorrent();
let info = SAMPLESITE;
console.log("adding", info);
client.add(info, torrent => {
  console.log("torrent metadata is ready");
  for (let file of torrent.files) {
    console.log("file:", file.name);
  }
});
