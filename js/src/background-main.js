// This code runs in the persistent background page.

import { createStore } from "redux";

import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

import reducers from "./reducers";

TrustedClient.init();

window.storage = new Storage();
window.store = createStore(reducers);
