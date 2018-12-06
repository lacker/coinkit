// This code runs in the persistent background page.
import LocalStorage from "./LocalStorage";
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

window.storage = new Storage(new LocalStorage());
TrustedClient.init(window.storage);

/*
// This changes the visible URL, which is unfortunate
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

// TODO: This doesn't seem to do anything
let config = {
  mode: "pac_script",
  pacScript: {
    data: "function FindProxyForURL(url, host) { return 'DIRECT'; }"
  }
};
chrome.proxy.settings.set({ value: config, scope: "regular" }, () => {
  console.log("proxy settings have been set:", config);
});
