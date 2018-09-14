// A screen to prompt a login.

import React, { Component } from "react";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";

import KeyPair from "./KeyPair";
import Styles from "./Styles";

export default class Login extends Component {
  // props.popup is a reference to the root popup
  constructor(props) {
    super(props);

    this.popup = props.popup;
    this.newAccount = this.newAccount.bind(this);

    this.state = {
      input: ""
    };
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
    this.popup.newKeyPair(kp);
    return true;
  }

  newAccount() {
    this.popup.newKeyPair(KeyPair.fromRandom());
  }

  // this.state.input could be a password or private key
  // TODO: handle password
  handleInput() {
    let kp = KeyPair.fromPrivateKey(this.state.input);
    if (kp) {
      this.popup.newKeyPair(kp);
    }

    // The input was not valid
    this.setState({ input: "" });
  }

  render() {
    return (
      <div style={Styles.popup}>
        <h1>Login</h1>
        <form
          onSubmit={event => {
            event.preventDefault();
            this.handleInput();
          }}
        >
          <div>Password or private key</div>
          <TextField
            type="password"
            value={this.state.input}
            autoFocus={true}
            onChange={event => {
              this.setState({
                input: event.target.value
              });
            }}
          />
          <Button variant="contained" color="primary" type="submit">
            Log In
          </Button>
        </form>
        <div
          style={{ color: "blue", cursor: "pointer" }}
          onClick={this.newAccount}
        >
          Create a new account
        </div>
      </div>
    );
  }
}
