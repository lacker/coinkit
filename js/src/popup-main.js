import React from "react";
import ReactDOM from "react-dom";
import "typeface-roboto";

import Popup from "./Popup";

// This code runs to load the popup of the chrome extension.

window.onload = () => {
  ReactDOM.render(<Popup />, document.getElementById("root"));
};
