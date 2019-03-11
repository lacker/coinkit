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
    let provider = null;

    // buckets maps bucket name to information about the bucket
    let buckets = {};

    while (true) {
      // XXX
    }
  }
}

module.exports = ProviderListener;
