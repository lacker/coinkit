// An integration test
const WebTorrent = require("webtorrent");
const TorrentClient = require("./TorrentClient.js");

const SINTEL =
  "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent";

const PIC =
  "magnet:?xt=urn:btih:4c4b94414ed2cd9b3b2db4c58006610746ceeda8&dn=DrfuEaUX0AUea-5.jpg&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com";

const BUNNY =
  "magnet:?xt=urn:btih:dd8255ecdc7ca55fb0bbf81323d87062db1f6d1c&dn=Big+Buck+Bunny&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fbig-buck-bunny.torrent";

async function main() {
  let client = new TorrentClient();
  let magnet = SINTEL;
  let torrent = client.download(magnet);
  await torrent.monitorProgress();
  client.destroy();
}

/* This one at least worked for sintel
async function main() {
  let client = new WebTorrent();

  let magnet = BUNNY;

  let promise = new Promise((resolve, reject) => {
    client.add(magnet, torrent => {
      for (let file of torrent.files) {
        console.log("found file with name:", file.name);
      }

      torrent.on("done", () => {
        console.log("downloaded", torrent.downloaded, "bytes");
        console.log("progress:", torrent.progress);
        client.destroy(err => {
          if (err) {
            console.log("error in destruction:", err);
          }
          resolve(null);
        });
      });
    });
  });

  return await promise;
}
*/

main()
  .then(() => {
    console.log("done");
  })
  .catch(e => {
    console.log("Unhandled " + e);
    process.exit(1);
  });
