// An integration test
const WebTorrent = require("webtorrent");

async function main() {
  let client = new WebTorrent();

  let magnet =
    "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent";

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

main()
  .then(() => {
    console.log("done");
  })
  .catch(e => {
    console.log("Unhandled " + e);
    process.exit(1);
  });
