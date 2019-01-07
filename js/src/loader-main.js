// This code is injected into .coinkit pages in order to load their actual content.

import TorrentClient from "./TorrentClient";

console.log("torrent-loading", window.location.href);
window.stop();

async function load() {
  let client = new TorrentClient();
  let parts = window.location.hostname.split(".");
  if (parts.length != 2 || parts[1] != "coinkit") {
    throw new Error("unexpected hostname: " + window.location.hostname);
  }
  let domain = parts[0];
  let pathname = window.location.pathname;
  if (pathname.endsWith("/")) {
    pathname += "index.html";
  }
  let html = await client.getAsText(domain, pathname);
  document.write(html);
}

load().catch(e => {
  console.log("loading error:", e);
});
