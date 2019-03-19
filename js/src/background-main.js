// This code runs in the persistent background page.
import LocalStorage from "./LocalStorage";
import Storage from "./Storage";
import TorrentDownloader from "./TorrentDownloader";
import TrustedClient from "./TrustedClient";

window.storage = new Storage(new LocalStorage());
TrustedClient.init(window.storage);

// Creates a pac script so that all .coinkit URLs get proxied to a
// black hole server.
//
// All that a "black hole server" needs to do is return a valid http
// response. It can be blank. It can be any other content, too, since
// the extension will stop all content loading and load the real site
// via the distributed system. So the content might as well be blank.
//
// We need to do this method for redirecting .coinkit domains so that
// the URL still appears as .coinkit in the browser. I think this
// necessary so that the behavior is comprehensible to the end user.
//
// This is not ideal architecturally. In particular, information on
// what URLs we are loading does get leaked to the proxy. And we are
// dependent on finding a usable proxy site. But I think the tradeoff
// is worth it for increased usability.
function buildBlackHoleScript(server) {
  let script = `
    function FindProxyForURL(url, host) {
      if (shExpMatch(host, "*.coinkit")) {
        return "PROXY ${server}";
      }
      return 'DIRECT';
    }
  `;
  return script;
}

// Update the black hole proxy
async function setBlackHoleProxy(server) {
  let script = buildBlackHoleScript(server);
  let config = {
    mode: "pac_script",
    pacScript: {
      data: script
    }
  };

  return await new Promise((resolve, reject) => {
    chrome.proxy.settings.set({ value: config, scope: "regular" }, () => {
      console.log("proxy settings updated. black hole is", server);
      resolve();
    });
  });
}

// For now there must be a black hole proxy running on localhost:3333.
// Later this proxy address will need to be loaded dynamically from somewhere.
setBlackHoleProxy("localhost:3333", {}).then(() => {
  console.log("initial black hole proxy configuration complete");
});

let downloader = new TorrentDownloader();

// Handle non-html requests by redirecting them to a data URL
chrome.webRequest.onBeforeRequest.addListener(
  details => {
    let url = new URL(details.url);
    let file = downloader.getFileFromCache(url.hostname, url.pathname);
    if (!file.data) {
      console.log("no data found for", url.hostname, url.pathname);
      return { redirectUrl: "about:blank" };
    }
    console.log("data found for", url.hostname, url.pathname);
    return { redirectUrl: file.data };
  },
  {
    urls: ["*://*.coinkit/*"],
    types: [
      "font",
      "image",
      "media",
      "object",
      "script",
      "stylesheet",
      "xmlhttprequest"
    ]
  },
  ["blocking"]
);

// Just logs completed coinkit navigation requests
chrome.webRequest.onCompleted.addListener(
  details => {
    let url = new URL(details.url);
    console.log("html request completed for", url.hostname, url.pathname);
  },
  {
    urls: ["*://*.coinkit/*"],
    types: ["main_frame", "sub_frame"]
  }
);

// Listen for the loader wanting a file
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (!message.getFile) {
    return false;
  }
  let { hostname, pathname } = message.getFile;
  downloader
    .getFile(hostname, pathname)
    .then(file => {
      // TODO: handle non html stuff
      console.log("sending response:", file.html);
      sendResponse(file.html);
    })
    .catch(e => {
      console.log("sending error response:", e);
      sendResponse({ error: e.message });
    });
  return true;
});
