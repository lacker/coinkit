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

  async remove(infoHash) {
    this.log("removing", infoHash);
    await this.client.remove(infoHash);
    // TODO: remove the actual file
  }

  async handleBuckets(buckets) {
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
        await this.remove(infoHash);
      }
    }

    // Handle data that is being added
    for (let infoHash in newInfoMap) {
      if (!this.infoMap[infoHash]) {
        this.log("adding", infoHash);

        // Start seeding this torrent. If the directory is already there from a previous run,
        // this should reuse it.
        let dir = this.subdirectory(infoHash);
        let bucket = newInfoMap[infoHash];
        let torrent = this.client.download(bucket.magnet, dir);

        // Check to make sure that this torrent isn't too large
        await torrent.waitForMetadata();
        let bucketBytes = bucket.size * 1024 * 1024;
        let torrentBytes = torrent.totalBytes();
        if (torrentBytes > bucketBytes) {
          // The torrent *is* too large.
          this.log(
            "torrent",
            infoHash,
            "contains",
            torrentBytes,
            "bytes but bucket",
            bucket.name,
            "only holds",
            bucketBytes,
            "bytes"
          );
          await this.remove(infoHash);
        }
      }
    }

    this.infoMap = newInfoMap;
  }

  async serve() {
    await this.listener.listen(this.id);
  }
}

module.exports = HostingServer;
