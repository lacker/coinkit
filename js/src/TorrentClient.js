// The TorrentClient finds the right torrent for a domain and uses that to return subsequent
// files.

// TODO: give this functions that lets it serve up subsequent files like image files

import WebTorrent from "webtorrent";

// The initial server that tells us where to start finding peers
let BOOTSTRAP = "http://localhost:4444";

class TorrentClient {
  constructor() {
    this.client = new WebTorrent();

    // Maps domain to {magnet, time} object
    this.magnets = {};

    // Maps domain to torrent
    this.torrents = {};
  }

  // Starts downloading and resolves to a torrent object when the download finishes
  async loadMagnet(magnet) {
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

  // Starts downloading a domain and resolves when the root URL is ready
  // Resolves to the torrent object
  // TODO: require a certain amount of domain newness
  async loadDomain(domain) {
    if (this.torrents[domain]) {
      return this.torrents[domain];
    }
    let magnet = await this.getMagnet(domain);
    let torrent = await this.loadMagnet(magnet);
    this.torrents[domain] = torrent;
    return torrent;
  }

  isReady(domain) {
    return domain in this.torrents;
  }

  // Rejects if there is no such file.
  async getAsText(domain, name) {
    let torrent = await this.loadDomain(domain);
    let file = torrent.files.find(file => file.name === name);
    if (!file) {
      return Promise.reject();
    }

    return new Promise((resolve, reject) => {
      file.getBlog((err, blob) => {
        if (err) {
          reject(err);
        }
        let reader = new FileReader();
        reader.onload = e => {
          resolve(e.target.result);
        };
        reader.readAsText(blob);
      });
    });
  }
}
