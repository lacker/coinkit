import React from "react";
import ReactDOM from "react-dom";
import "typeface-roboto";

import Popup from "./Popup";
import Storage from "./Storage";

// This code runs to load the popup of the chrome extension.
async function onload() {
  let storage = chrome.extension.getBackgroundPage().storage;
  if (!storage) {
    throw new Error("cannot find storage");
  }
  await storage.init();

  ReactDOM.render(<Popup />, document.getElementById("root"));
}

window.onload = onload;
