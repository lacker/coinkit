import React from "react";
import ReactDOM from "react-dom";

import App from "./App";

// This code runs at the root of our sample app.

window.onload = () => {
  ReactDOM.render(<App />, document.getElementById("root"));
};
