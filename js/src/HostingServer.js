// The node hosting server that miners run to store files.

const ProviderListener = require("./ProviderListener.js");

class HostingServer {
  // id is the provider id we are hosting for
  constructor(id) {
    this.id = id;
    this.verbose = false;

    this.listener = new ProviderListener();
    this.listener.onAdd(magnet => this.add(magnet));
    this.listener.onRemove(magnet => this.remove(magnet));
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
    }
  }

  add(magnet) {
    this.log("adding magnet:", magnet);
  }

  remove(magnet) {
    this.log("removing magnet:", magnet);
  }

  async serve() {
    await this.listener.listen(this.id);
  }
}

module.exports = HostingServer;
