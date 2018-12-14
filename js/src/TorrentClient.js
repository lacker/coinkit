// The TorrentClient finds the right torrent for a domain and uses that to return subsequent files.

// TODO: give this functions that lets it serve up subsequent files like image files

import WebTorrent from "webtorrent";

// The initial server that tells us where to start finding peers
let BOOTSTRAP = "http://localhost:4444";

class TorrentClient {
  constructor() {
    this.client = new WebTorrent();
  }

  // Starts downloading and resolves to a torrent object when the download finishes
  async download(magnet) {
    return new Promise((resolve, reject) => {
      this.client.add(magnet, torrent => {
        torrent.on("done", () => {
          resolve(torrent);
        });
      });
    });
  }

  // Returns a promise that maps to a magnet url
  async getMagnet(domain) {
    let response = await fetch(BOOTSTRAP);
    let json = await response.json();
    return json.magnet;
  }

  // TODO: describe what and when this returns
  async startDownloading(domain) {
    let magnet = this.getMagnet(domain);
    // TODO: more
  }
}
