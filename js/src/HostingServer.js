// The node hosting server that miners run to store files.

const ProviderListener = require("./ProviderListener.js");

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

    this.listener = new ProviderListener(verbose);
    this.listener.onAdd(magnet => this.add(magnet));
    this.listener.onRemove(magnet => this.remove(magnet));
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
    }
  }

  add(magnet) {
    let infoHash = getInfoHash(magnet);
    this.log("adding:", infoHash);
  }

  remove(magnet) {
    let infoHash = getInfoHash(magnet);
    this.log("removing:", infoHash);
  }

  async serve() {
    await this.listener.listen(this.id);
  }
}

module.exports = HostingServer;
