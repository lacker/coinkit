// This code runs in the persistent background page.
import LocalStorage from "./LocalStorage";
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

window.storage = new Storage(new LocalStorage());
TrustedClient.init(window.storage);

chrome.webRequest.onErrorOccurred.addListener(
  details => {
    let a = document.createElement("a");
    a.href = details.url;
    let parts = a.hostname.split(".");
    if (parts.length != 2) {
      return;
    }
    let [domain, tld] = parts;
    if (tld != "coinkit") {
      return;
    }

    console.log(
      "http request failed to domain:",
      domain,
      "on tab",
      details.tabId
    );
  },
  {
    urls: ["*://*.coinkit/*"],
    types: ["main_frame"]
  }
);

/*
// This changes the visible URL, which is bad for main_frame requests
chrome.webRequest.onBeforeRequest.addListener(
  details => {
    console.log("request initiated to", details.url);

    return { redirectUrl: "data:text/plain,hello world" };
  },
  {
    urls: ["*://*.coinkit/*"]
  },
  ["blocking"]
);
*/

/*
// This proxies .coinkit requests somewhere, which requires that we have a proxy
let script = `
function FindProxyForURL(url, host) {
  if (shExpMatch(host, "*.coinkit")) {
    return "PROXY about:blank";
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
*/
