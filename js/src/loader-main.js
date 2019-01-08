// This code is injected into .coinkit pages in order to load their actual content.

import TorrentClient from "./TorrentClient";

window.stop();

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
