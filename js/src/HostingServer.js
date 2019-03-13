// The node hosting server that miners run to store files.

// TODO: see if this is still needed. You may or may not first have to run:
// `sudo apt-get install xvfb`
// `sudo apt-get install libgconf-2-4`
// It used to be necessary but the webtorrent library may have gotten updated.

class HostingServer {
  // id is the provider id we are hosting for
  constructor(id) {
    this.id = id;
    this.verbose = false;
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
}

module.exports = HostingServer;
