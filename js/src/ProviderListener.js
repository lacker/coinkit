const ChainClient = require("./ChainClient.js");

// The ProviderListener continuously tracks information relevant to a single provider.
// This is designed to be the source of information for a hosting server.
class ProviderListener {
  constructor(id) {
    this.id = id;
    this.verbose = false;
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
    }
  }
}

module.exports = ProviderListener;
