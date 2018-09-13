// A screen to let the user create a password to locally encrypt keys.

import React, { Component } from "react";
import TextField from "@material-ui/core/TextField";

export default class NewPassword extends Component {
  // props.popup is a reference to the root popup
  constructor(props) {
    super(props);

    this.popup = props.popup;

    this.state = {
      password: "",
      repeatPassword: ""
    };
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

    let warning = "";
    if (
      this.state.password != "" &&
      this.state.repeatPassword != "" &&
      this.state.password != this.state.repeatPassword
    ) {
      warning = "passwords must match";
    }

    return (
      <div style={style}>
        <h1>Choose a password</h1>
        <div>Password</div>
        <TextField autofocus={true} value={this.state.password} />
        <div>Repeat your password</div>
        <TextField value={this.state.repeatPassword} />
        <div>{warning}</div>
      </div>
    );
  }
}
