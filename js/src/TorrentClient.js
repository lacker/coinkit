// A wrapper around the WebTorrent client with an async API.
const WebTorrent = require("webtorrent");

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
}

module.exports = TorrentClient;
