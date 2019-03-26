const path = require("path");

const TorrentClient = require("./TorrentClient.js");

test("Creating and shutting down a torrent client", async () => {
  let client = new TorrentClient();
  await client.destroy();
});

test(
  "Seeding a torrent, then downloading it",
  async () => {
    // Seed a torrent
    let seedClient = new TorrentClient();
    let dir = path.resolve(__dirname, "samplesite");
    let t = await seedClient.seed(dir);

    // Download that exact torrent
    let downloadClient = new TorrentClient();
    let torrent = await downloadClient.download(t.magnet);
    await torrent.waitForDone();

    // TODO: check the data is ok

    await seedClient.destroy();
    await downloadClient.destroy();
  },
  30000
);
