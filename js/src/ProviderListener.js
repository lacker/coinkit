const ChainClient = require("./ChainClient.js");
const { sleep } = require("./Util.js");

// The ProviderListener continuously tracks information relevant to a single provider.
// This is designed to be the source of information for a hosting server.
class ProviderListener {
  constructor() {
    this.client = new ChainClient();
    this.verbose = false;
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
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
          }
        } else {
          this.log("allocate bucket:", bucket);
        }
        newBuckets[bucket.name] = bucket;
      }

      // Check for dropped buckets
      for (let name in newBuckets) {
        if (!buckets[name]) {
          this.log("deallocate bucket:", name);
        }
      }

      await sleep(1000);
    }
  }
}

module.exports = ProviderListener;
