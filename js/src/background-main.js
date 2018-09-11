// This code runs in the persistent background page.
import Storage from "./Storage";

// TODO: Access this from the popup with chrome.extension.getBackgroundPage().storage
let storage = new Storage();
