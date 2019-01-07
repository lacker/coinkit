// This code is injected into .coinkit pages in order to load their actual content.

import TorrentClient from "./TorrentClient";

console.log("torrent-loading", window.location.href);
window.stop();

async function load() {
  let client = new TorrentClient();
  let html = await client.getFile(
    window.location.hostname,
    window.location.pathname
  );
  document.write(html);
}

load().catch(e => {
  console.log("loading error:", e);
});
