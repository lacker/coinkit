// This code runs in the persistent background page.
import Storage from "./Storage";

// Access this from the popup with chrome.extension.getBackgroundPage().storage
window.storage = new Storage();
