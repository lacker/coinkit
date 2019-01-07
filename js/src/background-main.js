// This code runs in the persistent background page.
import LocalStorage from "./LocalStorage";
import Storage from "./Storage";
import TorrentClient from "./TorrentClient";
import TrustedClient from "./TrustedClient";

window.storage = new Storage(new LocalStorage());
TrustedClient.init(window.storage);

// This proxies .coinkit requests somewhere, which requires that we have a proxy.
// For now let's assume there is a proxy running on localhost:3333.
// Later this proxy address will need to be loaded dynamically from somewhere.
let script = `
function FindProxyForURL(url, host) {
  if (shExpMatch(host, "*.coinkit")) {
    return "PROXY localhost:3333";
  }
  return 'DIRECT';
}
`;

let config = {
  mode: "pac_script",
  pacScript: {
    data: script
  }
};
chrome.proxy.settings.set({ value: config, scope: "regular" }, () => {
  console.log("proxy settings have been set:", config);
});

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
