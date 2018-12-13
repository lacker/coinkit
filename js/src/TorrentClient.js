// The TorrentClient finds the right torrent for a domain and uses that to return subsequent files.

import WebTorrent from "webtorrent";

class TorrentClient {
  constructor() {
    this.client = new WebTorrent();
  }
}
