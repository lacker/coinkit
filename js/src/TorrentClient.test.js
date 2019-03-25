const path = require("path");

const TorrentClient = require("./TorrentClient.js");

test("Seeding and downloading", async () => {
  let client = new TorrentClient();
  let dir = path.resolve(__dirname, "samplesite");
});
