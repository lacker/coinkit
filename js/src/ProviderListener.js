const ChainClient = require("./ChainClient.js");
const { sleep } = require("./Util.js");

// The ProviderListener continuously tracks information relevant to a single provider.
// This is designed to be the source of information for a hosting server.
class ProviderListener {
  constructor(verbose) {
    this.client = new ChainClient();
    this.verbose = !!verbose;
    this.addCallback = null;
    this.removeCallback = null;

    // Number of update cycles this listener has gone through
    this.updates = 0;
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
    }
  }

  onAdd(f) {
    this.addCallback = f;
  }

  onRemove(f) {
    this.removeCallback = f;
  }

  handleAdd(magnet) {
    if (this.addCallback) {
      this.addCallback(magnet);
    }
  }

  handleRemove(magnet) {
    if (this.removeCallback) {
      this.removeCallback(magnet);
    }
  }

  // Listens forever
  async listen(id) {
    // buckets maps bucket name to information about the bucket
    let buckets = {};

    while (true) {
      let bucketList = await this.client.getBuckets({ provider: id });

      let newBuckets = {};
      for (let bucket of bucketList) {
        let oldVersion = buckets[bucket.name];
        if (oldVersion) {
          if (oldVersion.magnet != bucket.magnet) {
            this.log(bucket.name, "bucket has new magnet:", bucket.magnet);
            this.handleRemove(oldVersion.magnet);
            this.handleAdd(bucket.magnet);
          }
        } else {
          this.log("allocate bucket:", bucket.name);
          this.handleAdd(bucket.magnet);
        }
        newBuckets[bucket.name] = bucket;
      }

      // Check for dropped buckets
      for (let name in buckets) {
        if (!newBuckets[name]) {
          this.log("deallocate bucket:", name);
          this.handleRemove(buckets[name].magnet);
        }
      }

      buckets = newBuckets;
      this.updates += 1;
      await sleep(1000);
    }
  }
}

module.exports = ProviderListener;
