import React from "react";
import ReactDOM from "react-dom";
import { Provider } from "react-redux";
import "typeface-roboto";

import Popup from "./Popup";
import Storage from "./Storage";

// This code runs to load the popup of the chrome extension.
async function onload() {
  let store = chrome.extension.getBackgroundPage().store;
  if (!store) {
    throw new Error("cannot find store");
  }

  console.log("XXX initial state:", store.getState());

  ReactDOM.render(
    <Provider store={store}>
      <Popup />
    </Provider>,
    document.getElementById("root")
  );
}

window.onload = onload;
