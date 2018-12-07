// This code runs in the persistent background page.
import LocalStorage from "./LocalStorage";
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

window.storage = new Storage(new LocalStorage());
TrustedClient.init(window.storage);

// This proxies .coinkit requests somewhere, which requires that we have a proxy.
// For now let's assume there is a proxy running on localhost:3333.

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
