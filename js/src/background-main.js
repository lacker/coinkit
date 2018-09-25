// This code runs in the persistent background page.

import { createStore } from "redux";

import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

import reducers from "./reducers";

window.client = new TrustedClient();
window.storage = new Storage();
window.store = createStore(reducers);
