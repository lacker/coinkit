// A wrapper around the WebTorrent client with an async API.
const WebTorrent = require("webtorrent");

const Torrent = require("./Torrent.js");

class TorrentClient {
  constructor() {
    this.client = new WebTorrent();
    this.client.on("error", err => {
      console.log("fatal error:", err.message);
    });
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

  // Returns a Torrent object for downloading this magnet url.
  // Does not wait for the download to complete before returning.
  // If you want that, call waitForDone.
  async download(magnet) {
    // First, check if this download is already in progress.
    for (let t of this.client.torrents) {
      if (t.magnetURI == magnet) {
        return new Torrent(t);
      }
    }

    // Add a new download
    let t = this.client.add(magnet);
    return new Torrent(t);
  }

  // Shuts down the torrent client.
  async destroy() {
    let promise = new Promise((resolve, reject) => {
      this.client.destroy(err => {
        resolve(null);
      });
    });
    return await promise;
  }
}

module.exports = TorrentClient;
