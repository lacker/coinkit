const path = require("path");

const TorrentClient = require("./TorrentClient.js");

test("Creating and shutting down a torrent client", async () => {
  let client = new TorrentClient();
  await client.destroy();
});

// TODO: neither of these tests work. TorrentClient is broken in some unknown way.

test.skip("Downloading a test torrent", async () => {
  /*
  let magnet =
    "magnet:?xt=urn:btih:209c8226b299b308beaf2b9cd3fb49212dbd13ec&dn=Tears+of+Steel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Ftears-of-steel.torrent";
  */

  let magnet =
    "magnet:?xt=urn:btih:dd8255ecdc7ca55fb0bbf81323d87062db1f6d1c&dn=Big+Buck+Bunny&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fbig-buck-bunny.torrent";

  let client = new TorrentClient();
  client.verbose = true;
  let torrent = client.download(magnet);
  await torrent.waitForDone();
  // TODO: check something about the data

  await client.destroy();
});

test.skip(
  "Seeding a torrent, then downloading it",
  async () => {
    // Seed a torrent
    let seedClient = new TorrentClient();
    let dir = path.resolve(__dirname, "samplesite");
    console.log("test: waiting for seed");
    let t = await seedClient.seed(dir);

    // Download that exact torrent
    let downloadClient = new TorrentClient();
    let torrent = await downloadClient.download(t.magnet);
    torrent.verbose = true;

    console.log("test: waitForDone");
    await torrent.waitForDone();
    console.log("test: the promised land!");

    // TODO: check the data is ok

    await seedClient.destroy();
    await downloadClient.destroy();
  },
  30000
);
