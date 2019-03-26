const path = require("path");

const TorrentClient = require("./TorrentClient.js");

test("Creating and shutting down a torrent client", async () => {
  let client = new TorrentClient();
  await client.destroy();
});

test(
  "Seeding a torrent",
  async () => {
    let seedClient = new TorrentClient();
    let dir = path.resolve(__dirname, "samplesite");
    let t = await seedClient.seed(dir);
    await seedClient.destroy();
  },
  10000
);

/*
  // Download that exact torrent
  let downloadClient = new TorrentClient();
  let torrent = await downloadClient.download(magnet);
  await torrent.waitForDone();

  // TODO: check that the download works
});

*/
