// A wrapper around WebTorrent's "torrent" object with an async API.
const { sleep } = require("./Util.js");

class Torrent {
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
}

module.exports = Torrent;
