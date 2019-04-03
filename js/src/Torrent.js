// A wrapper around WebTorrent's "torrent" object with an async API.
const { sleep } = require("./Util.js");

class Torrent {
  // This constructor should be cheap, since we often construct many Torrent objects from
  // the same underlying torrent.
  constructor(torrent, verbose) {
    this.torrent = torrent;
    this.magnet = torrent.magnetURI;
    this.infoHash = torrent.infoHash;
    this.verbose = !!verbose;
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
    }
  }

  isDone() {
    return this.torrent.progress == 1;
  }

  async monitorProgress() {
    while (!this.isDone()) {
      console.log("progress:", this.torrent.progress);
      await sleep(1000);
    }
    console.log(
      "progress complete.",
      this.torrent.downloaded,
      "bytes downloaded"
    );
  }

  // If you call this before metadata it will just return 0
  totalBytes() {
    let answer = 0;
    for (let file of this.torrent.files) {
      answer += file.length;
    }
    return answer;
  }

  // Always returns null
  async waitForMetadata() {
    if (this.torrent.files.length > 0) {
      return null;
    }

    return new Promise((resolve, reject) => {
      this.torrent.on("metadata", () => {
        resolve(null);
      });
    });
  }

  // Always returns null
  async waitForDone() {
    this.log("progress:", this.torrent.progress);
    if (this.isDone()) {
      this.log("waitForDone is done because we are already done");
      return null;
    }
    let promise = new Promise((resolve, reject) => {
      this.log("waiting for 'done' event");
      this.torrent.on("done", () => {
        resolve(null);
      });
    });
    return await promise;
  }

  // Waits until there are n seeds for this torrent
  async waitForSeeds(n) {
    // TODO: do the right thing here, instead of the wrong thing
    while (true) {
      await sleep(1000);
    }
  }

  // Shuts down this torrent
  async destroy() {
    let promise = new Promise((resolve, reject) => {
      this.torrent.destroy(() => {
        resolve(null);
      });
    });
    return await promise;
  }

  // Returns a map from filename to data
  async readAll() {
    // TODO
  }
}

module.exports = Torrent;
