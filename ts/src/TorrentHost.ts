// The TorrentHost hosts webtorrents, possibly over a long period of time, running
// on node, backed by the filesystem.

const TorrentClient = require("./TorrentClient.js");

class TorrentHost {
  constructor() {
    this.client = new TorrentClient();
  }
}

module.exports = TorrentHost;
