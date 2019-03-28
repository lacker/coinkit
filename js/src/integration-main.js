// An integration test

const path = require("path");

const TorrentClient = require("./TorrentClient.js");

async function main() {
  console.log("seeding...");
  let seedClient = new TorrentClient();
  let dir = path.resolve(__dirname, "samplesite");
  let t = await seedClient.seed(dir);
  let magnet = t.magnet;

  console.log("downloading...");
  let downloadClient = new TorrentClient();
  let torrent = await downloadClient.download(magnet);
  await torrent.monitorProgress();

  console.log("cleaning up...");
  await seedClient.destroy();
  await downloadClient.destroy();
}

main().then(() => {
  console.log(
    "done with integration test code. waiting for timer autocleanup..."
  );
});
