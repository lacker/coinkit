// The node hosting server that miners run to store files.

const path = require("path");

const ProviderListener = require("./ProviderListener.js");
const TorrentClient = require("./TorrentClient.js");

// Throws an error if the magnet url is an unknown format
function getInfoHash(magnet) {
  let prefix = "magnet:?xt=urn:btih:";
  if (!magnet.startsWith(prefix)) {
    throw new Error("unknown magnet format: " + magnet);
  }

  let rest = magnet.replace(prefix, "");
  return rest.split("&")[0];
}

class HostingServer {
  // id is the provider id we are hosting for
  constructor(id, directory, verbose) {
    this.id = id;
    this.directory = directory;
    this.verbose = !!verbose;
    this.client = new TorrentClient(this.verbose);

    // Maps info hash to bucket object
    this.infoMap = {};

    this.listener = new ProviderListener(verbose);
    this.listener.onBuckets(buckets => this.handleBuckets(buckets));
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
    }
  }

  subdirectory(infoHash) {
    return path.join(this.directory, infoHash);
  }

  handleBuckets(buckets) {
    // Figure out the new info map
    let newInfoMap = {};
    for (let bucket of buckets) {
      let infoHash;
      try {
        infoHash = getInfoHash(bucket.magnet);
      } catch (e) {
        console.log(e.message);
        continue;
      }
      newInfoMap[infoHash] = bucket;
    }

    // Handle data that is being deleted
    for (let infoHash in this.infoMap) {
      if (!newInfoMap[infoHash]) {
        this.log("removing:", infoHash);
        // TODO: clear this directory and remove the torrent from our client
      }
    }

    // Handle data that is being added
    for (let infoHash in newInfoMap) {
      if (!this.infoMap[infoHash]) {
        this.log("adding:", infoHash);

        // Start seeding this torrent. If the directory is already there from a previous run,
        // this should reuse it.
        let dir = this.subdirectory(infoHash);
        let bucket = newInfoMap[infoHash];
        this.client.add(bucket.magnet, { path: dir });
      }
    }

    this.infoMap = newInfoMap;
  }

  async serve() {
    await this.listener.listen(this.id);
  }
}

module.exports = HostingServer;
