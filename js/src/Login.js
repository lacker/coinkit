// A screen to prompt a login.

import React, { Component } from "react";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";

import KeyPair from "./KeyPair";

export default class Login extends Component {
  // props.popup is a reference to the root popup
  constructor(props) {
    super(props);

    this.popup = props.popup;
  }

  // Returns whether the private key is valid.
  // If it is valid, calls the callback on the associated keypair.
  setPrivateKey(privateKey) {
    let kp = null;
    try {
      kp = KeyPair.fromPrivateKey(privateKey);
    } catch (e) {
      return false;
    }
    this.popup.setKeyPair(kp);
    return true;
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
        <h1>Welcome</h1>
        <div>Password or private key</div>
        <TextField />
        <div>Create a new account</div>
      </div>
    );
  }
}
