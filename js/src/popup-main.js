import React from "react";
import ReactDOM from "react-dom";
import { Provider } from "react-redux";
import { createStore } from "redux";
import "typeface-roboto";

import Popup from "./Popup";
import Storage from "./Storage";
import { loadFromStorage } from "./actions";
import reducers from "./reducers";

// This code runs to load the popup of the chrome extension.
async function onload() {
  // Each popup gets its own redux store object.
  // I tried to let them share one but ran into weird bugs.
  let store = createStore(reducers);
  let storage = await Storage.get();
  store.dispatch(loadFromStorage(storage));

  // Save all state updates when there is a password set to retrieve them
  store.subscribe(() => {
    let state = store.getState();
    if (state.password == null && state.keyPair == null) {
      storage.logOut();
    } else if (state.password != null) {
      storage.setPasswordAndData(
        state.password,
        state.keyPair,
        state.permissions
      );
    }
  });

  ReactDOM.render(
    <Provider store={store}>
      <Popup />
    </Provider>,
    document.getElementById("root")
  );
}

window.onload = onload;
