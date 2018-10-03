// This code runs in the persistent background page.
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

window.storage = new Storage();
TrustedClient.init(window.storage);
