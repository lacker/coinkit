// The node hosting server that miners run to store files.

const fs = require("fs");
const path = require("path");

const rimraf = require("rimraf");

const ChainClient = require("./ChainClient.js");
const ProviderListener = require("./ProviderListener.js");
const TorrentClient = require("./TorrentClient.js");
const { isDirectory, loadKeyPair } = require("./FileUtil.js");

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
  // options must contain exactly one way to specify the provider:
  // id - the id of the provider
  // keyPair - the filename containing keys for the owner
  // other options:
  // capacity - how much space we have to store files, in megabytes
  // directory - where to store the hosted files
  // verbose - defaults to false
  constructor(options) {
    if (options.id && options.keyPair) {
      throw new Error(
        "only one of the id and keyPair options can be set for HostingServer"
      );
    }

    if (!isDirectory(options.directory)) {
      throw new Error(options.directory + " is not a directory");
    } else {
      this.log("hosting files in", options.directory);
    }

    if (options.keyPair) {
      this.keyPair = loadKeyPair(options.keyPair);
    }

    this.capacity = options.capacity;
    this.id = options.id;
    this.directory = options.directory;
    this.verbose = !!options.verbose;
    this.client = new TorrentClient(this.verbose);

    // Maps info hash to bucket object
    this.infoMap = {};

    this.listener = new ProviderListener(this.verbose);
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

  // Also cleans up the files on disk
  async remove(infoHash) {
    if (infoHash.length < 5) {
      this.log("infoHash suspiciously short:", infoHash);
      return null;
    }
    this.log("removing", infoHash);
    await this.client.remove(infoHash);
    let promise = new Promise((resolve, reject) => {
      rimraf(this.subdirectory(infoHash), { disableGlob: true }, err => {
        if (err) {
          this.log("rimraf error:", err.message);
        }
        resolve(null);
      });
    });
    return await promise;
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

  // Makes sure that this.id is set, creating a new provider if need be.
  async acquireProviderID() {
    if (this.id) {
      return this.id;
    }

    // Check to see if we already have a provider created.
    let client = new ChainClient(this.keyPair);
    let providers = await client.getProviders({
      owner: this.keyPair.getPublicKey()
    });
    if (providers.length > 1) {
      throw new Error(
        "HostingServer doesn't know how to handle a user that owns multiple providers"
      );
    }
    if (providers.length == 1) {
      let provider = providers[0];
      if (provider.capacity > this.capacity) {
        throw new Error(
          "the chain says we need " +
            provider.capacity +
            " capacity but we only have " +
            this.capacity
        );
      }
      this.id = provider.id;
      this.log("found provider id:", this.id);
      return this.id;
    }

    // Create a provider
    let provider = await client.createProvider(this.capacity);
    this.id = provider.id;
    this.log("created a new provider with id:", this.id);
    return this.id;
  }

  async serve() {
    try {
      await this.acquireProviderID();
    } catch (e) {
      console.log("failed to acquire provider id: " + e.message);
      process.exit(1);
    }
    await this.listener.listen(this.id);
  }
}

module.exports = HostingServer;
