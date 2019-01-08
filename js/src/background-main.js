// This code runs in the persistent background page.
import LocalStorage from "./LocalStorage";
import Storage from "./Storage";
import TorrentClient from "./TorrentClient";
import TrustedClient from "./TrustedClient";

window.storage = new Storage(new LocalStorage());
TrustedClient.init(window.storage);

// Creates a pac script so that:
// staticMap defines a map that sends URL to static content
// All other URLs ending in .coinkit get proxied to the provided proxy server
// All the proxy server needs to do is return a valid http response. It can be blank.
// It can be any other content, too, since the extension will stop it loading.
// But it might as well be blank.
function buildScript(server, staticMap) {
  // TODO: handle staticMap
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

// Sets the browser's proxy configuration script
async function setProxy(server, staticMap) {
  let script = buildScript(server, staticMap);
  let config = {
    mode: "pac_script",
    pacScript: {
      data: script
    }
  };

  return await new Promise((resolve, reject) => {
    chrome.proxy.settings.set({ value: config, scope: "regular" }, () => {
      console.log(
        "proxy settings updated. black hole is",
        server,
        "with static content for",
        Object.keys(staticMap).join(", ")
      );
      resolve();
    });
  });
}

// For now let's assume there is a proxy running on localhost:3333.
// Later this proxy address will need to be loaded dynamically from somewhere.
setProxy("localhost:3333", {}).then(() => {
  console.log("initial proxy configuration complete");
});

chrome.webRequest.onBeforeRequest.addListener(
  details => {
    console.log("request initiated for:", details.url);
  },
  {
    urls: ["*://*.coinkit/*"]
  },
  ["blocking"]
);

// Just logs completed coinkit requests
chrome.webRequest.onCompleted.addListener(
  details => {
    console.log(details.responseHeaders);

    let a = document.createElement("a");
    a.href = details.url;
    let parts = a.hostname.split(".");
    let tld = parts.pop();
    let domain = parts.pop();
    console.log("request completed for:", domain + "." + tld);
  },
  {
    urls: ["*://*.coinkit/*"],
    types: ["main_frame"]
  }
);

let client = new TorrentClient();

// Listen for the loader wanting a file
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (!message.getFile) {
    return false;
  }
  let { hostname, pathname } = message.getFile;
  client.getFile(hostname, pathname).then(file => {
    console.log("sending response:", file);
    sendResponse(file);
  });
  return true;
});
