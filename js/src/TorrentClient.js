// The TorrentClient finds the right torrent for a domain and uses that to return
// subsequent files.

import WebTorrent from "webtorrent";

// The initial server that tells us where to start finding peers
let BOOTSTRAP = "http://localhost:4444";

// Removes a leading / and adds a trailing index.html if needed
// so that callers can be indifferent
function cleanPathname(pathname) {
  if (pathname.endsWith("/")) {
    pathname += "index.html";
  }
  if (pathname.charAt(0) === "/") {
    pathname = pathname.substr(1);
  }
  return pathname;
}

// Async file reader
async function readFile(file) {
  return await new Promise((resolve, reject) => {
    file.getBlob((err, blob) => {
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

// Turns a torrent into a map from filename to text data
async function readTorrent(torrent) {
  let data = {};
  for (let file of torrent.files) {
    data[file.name] = await readFile(file);
  }
  return data;
}

// Adds a magnet to a WebTorrent client and resolves when it finishes
async function downloadTorrent(client, magnet) {
  return await new Promise((resolve, reject) => {
    client.add(magnet, torrent => {
      torrent.on("done", () => {
        resolve(torrent);
      });
    });
  });
}

export default class TorrentClient {
  constructor() {
    this.client = new WebTorrent();

    // Maps domain to {magnet, time} object
    this.magnets = {};

    // this.cache[magnet][filename] is the text cache for the file
    this.cache = {};
  }

  // Starts downloading and resolves when the download finishes.
  // Resolves to a map of filename to content, which is also stored in this.cache.
  async downloadMagnet(magnet) {
    if (this.cache[magnet]) {
      return this.cache[magnet];
    }
    let torrent = await downloadTorrent(this.client, magnet);
    let data = await readTorrent(torrent);
    this.cache[magnet] = data;
    return data;
  }

  // Returns a promise that maps to a magnet url
  // TODO: sometimes read from cache instead of just writing to it, have staleness logic
  async getMagnetURL(domain) {
    let response = await fetch(BOOTSTRAP);
    let json = await response.json();
    let magnet = json.magnet;
    let now = new Date();
    this.magnets[domain] = {
      magnet,
      time: now
    };
    return magnet;
  }

  // Starts downloading a domain and resolves when all files are ready
  // Resolves to a map from filename to content
  async downloadDomain(domain) {
    let magnet = await this.getMagnetURL(domain);
    return await this.downloadMagnet(magnet);
  }

  // Returns null if the file is not in the cache.
  getFileFromCache(domain, pathname) {
    pathname = cleanPathname(pathname);

    let magnetData = this.magnets[domain];
    if (!magnetData) {
      return null;
    }
    let magnet = magnetData.magnet;
    let cache = this.cache[magnet];
    if (!cache) {
      return null;
    }

    return cache[pathname] || null;
  }

  // Rejects if there is no such file.
  async getFile(domain, pathname) {
    pathname = cleanPathname(pathname);
    console.log("loading", pathname, "from the", domain, "domain");

    let data = await this.downloadDomain(domain);
    return data[pathname];
  }
}
