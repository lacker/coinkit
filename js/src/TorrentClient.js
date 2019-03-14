// A wrapper around the WebTorrent client with an async API.
const WebTorrent = require("webtorrent");
const { sleep } = require("./Util.js");

class TorrentClient {
  constructor() {
    this.client = new WebTorrent();
  }

  // Returns the WebTorrent object
  async seed(directory) {
    let promise = new Promise((resolve, reject) => {
      this.client.seed(directory, torrent => {
        resolve(torrent);
      });
    });
    return await promise;
  }

  // Waits for the provided torrent object to be seeded
  async waitForSeed(torrent) {
    // TODO: actually wait for the seed
    while (true) {
      await sleep(10000);
    }
  }
}

module.exports = TorrentClient;
