// This code runs in the persistent background page.

import { createStore } from "redux";

import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

import { loadFromStorage } from "./actions";
import reducers from "./reducers";

TrustedClient.init();

window.store = createStore(reducers);
window.storage = new Storage();

// Save all state updates when there is a password set to retrieve them
window.store.subscribe(() => {
  let state = window.store.getState();
  if (state.password != null) {
    // console.log("XXX saving state:", state);
    let data = {
      keyPair: state.keyPair.serialize()
    };
    window.storage.setPasswordAndData(state.password, data);
  }
});
