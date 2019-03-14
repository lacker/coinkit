// A wrapper around the WebTorrent client with an async API.
const WebTorrent = require("webtorrent");

const Torrent = require("./Torrent.js");

class TorrentClient {
  constructor() {
    this.client = new WebTorrent();
  }

  // Returns a Torrent object
  async seed(directory) {
    let promise = new Promise((resolve, reject) => {
      this.client.seed(directory, torrent => {
        resolve(new Torrent(torrent));
      });
    });
    return await promise;
  }
}

module.exports = TorrentClient;
