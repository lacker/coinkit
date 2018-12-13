// The TorrentClient finds the right torrent for a domain and uses that to return subsequent files.

// TODO: give this functions that lets it serve up subsequent files like image files

import WebTorrent from "webtorrent";

class TorrentClient {
  constructor() {
    this.client = new WebTorrent();
  }
}
