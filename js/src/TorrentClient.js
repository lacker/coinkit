// A wrapper around the WebTorrent client with an async API.
const WebTorrent = require("webtorrent-hybrid");

const Torrent = require("./Torrent.js");

const TRACKERS = ["ws://localhost:4444"];

function nicePeerId(id) {
  return "_" + ("" + id).slice(-4);
}

class TorrentClient {
  constructor(verbose) {
    this.client = new WebTorrent();
    this.client.on("error", err => {
      console.log("fatal error in TorrentClient:", err.message);
    });
    this.verbose = !!verbose;
    this.log("creating torrent client", nicePeerId(this.client.peerId));
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
    }
  }

  // Call on a raw WebTorrent object, not a wrapped Torrent
  logTorrentEvents(torrent) {
    torrent.on("metadata", () => {
      this.log("metadata acquired for", torrent.magnetURI);
    });
    torrent.on("warning", err => {
      this.log("warning:", err.message);
    });
    torrent.on("wire", (wire, addr) => {
      this.log("connected to", nicePeerId(wire.peerId), "at", addr);
    });
    torrent.on("error", err => {
      this.log("torrent error:", err.message);
    });
  }

  // Returns an array of Torrent objects
  getTorrents() {
    let answer = [];
    for (let t of this.client.torrents) {
      answer.push(new Torrent(t));
    }
    return answer;
  }

  // Returns a Torrent object
  async seed(directory) {
    let promise = new Promise((resolve, reject) => {
      this.client.seed(
        directory,
        {
          announceList: [TRACKERS]
        },
        torrent => {
          this.logTorrentEvents(torrent);
          resolve(new Torrent(torrent, this.verbose));
        }
      );
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
        return new Torrent(t, this.verbose);
      }
    }

    // Add a new download
    let t = this.client.add(magnet);
    this.log("downloading", magnet);
    this.logTorrentEvents(t);
    return new Torrent(t, this.verbose);
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
