// An integration test
const path = require("path");
const TorrentClient = require("./TorrentClient.js");

async function main() {
  // I seeded a random image torrent with instant.io and this is the magnet url
  let magnet =
    "magnet:?xt=urn:btih:4c4b94414ed2cd9b3b2db4c58006610746ceeda8&dn=DrfuEaUX0AUea-5.jpg&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com";

  let client = new TorrentClient();
  console.log("torrent client created");
  let torrent = await client.download(magnet);
  console.log("download has begun");
  await torrent.waitForDone();
  console.log("download complete");

  // TODO: check something about the data

  await client.destroy();
}

main()
  .then(() => {
    console.log("done");
  })
  .catch(e => {
    console.log("Unhandled " + e);
    process.exit(1);
  });
