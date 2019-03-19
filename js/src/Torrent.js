// A wrapper around WebTorrent's "torrent" object with an async API.
const WebTorrent = require("webtorrent");

const { sleep } = require("./Util.js");

class Torrent {
  constructor(torrent) {
    this.torrent = torrent;
    this.magnet = torrent.magnetURI;
    this.infoHash = torrent.infoHash;
  }

  isDone() {
    return this.torrent.progress == 1;
  }

  async waitForDone() {
    if (this.isDone()) {
      return null;
    }
    let promise = new Promise((resolve, reject) => {
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
}

module.exports = Torrent;
