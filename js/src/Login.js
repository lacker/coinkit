// A screen to prompt a login.

import React, { Component } from "react";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";

import KeyPair from "./KeyPair";

export default class Login extends Component {
  // props.callback takes a keypair once the user has logged in
  constructor(props) {
    super(props);

    this.callback = props.callback;
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
    this.callback(kp);
    return true;
  }

  render() {
    return (
      <div>
        <TextField>foo</TextField>
      </div>
    );
  }
}
