// A screen to let the user create a password to locally encrypt keys.

import React, { Component } from "react";
import TextField from "@material-ui/core/TextField";

export default class NewPassword extends Component {
  // props.popup is a reference to the root popup
  constructor(props) {
    super(props);

    this.popup = props.popup;
  }

  render() {
    let style = {
      display: "flex",
      flexDirection: "column",
      justifyContent: "flex-start",
      alignItems: "center",
      width: 360,
      padding: 30
    };

    return (
      <div style={style}>
        <h1>Choose a password</h1>
        <div>Password</div>
        <TextField />
        <div>Repeat your password</div>
        <TextField />
      </div>
    );
  }
}
