// This code is injected into .coinkit pages in order to load their actual content.

import TorrentClient from "./TorrentClient";

console.log("torrent-loading", window.location.href);
window.stop();

console.log("extension id is", chrome.runtime.id);

chrome.runtime.sendMessage({ getFile: "hello" }, response => {
  console.log("loader-main got response:", response);
});

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
