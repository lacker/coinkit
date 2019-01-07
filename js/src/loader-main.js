// This code is injected into .coinkit pages in order to load their actual content.

import TorrentClient from "./TorrentClient";

console.log("torrent-loading", window.location.href);
window.stop();

console.log("extension id is", chrome.runtime.id);

chrome.runtime.sendMessage(
  {
    getFile: {
      hostname: window.location.hostname,
      pathname: window.location.pathname
    }
  },
  response => {
    document.write(response);
  }
);
