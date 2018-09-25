// This code runs in the persistent background page.

import { createStore } from "redux";

import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

import { loadFromStorage } from "./actions";
import reducers from "./reducers";

TrustedClient.init();

window.store = createStore(reducers);
window.storage = new Storage();
