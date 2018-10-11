// This code runs in the persistent background page.
import LocalStorage from "./LocalStorage";
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

window.storage = new Storage(new LocalStorage());
TrustedClient.init(window.storage);
