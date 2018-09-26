// This code runs in the persistent background page.
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

TrustedClient.init();
window.storage = new Storage();
