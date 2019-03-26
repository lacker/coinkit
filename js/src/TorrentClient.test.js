const path = require("path");

const TorrentClient = require("./TorrentClient.js");

test("Creating a torrent client", async () => {
  let client = new TorrentClient();
});

/*

test("Seeding and downloading", async () => {
  // Seed a torrent
  let seedClient = new TorrentClient();
  let dir = path.resolve(__dirname, "samplesite");
  let t = await seedClient.seed(dir);
  let magnet = t.magnet;

  // Download that exact torrent
  let downloadClient = new TorrentClient();
  let torrent = await downloadClient.download(magnet);
  await torrent.waitForDone();

  // TODO: check that the download works
});

*/
