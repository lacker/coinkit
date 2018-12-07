// This code runs in the persistent background page.
import LocalStorage from "./LocalStorage";
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

window.storage = new Storage(new LocalStorage());
TrustedClient.init(window.storage);

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

let script = `
function FindProxyForURL(url, host) {
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
