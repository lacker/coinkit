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

  // this.state.input could be a password or private key
  // TODO: handle password
  handleInput() {
    // Check if the input was a private key
    let kp = KeyPair.fromPrivateKey(this.state.input);
    if (kp) {
      this.popup.newKeyPair(kp);
      return;
    }

    // Check if the input was a password
    this.popup.checkPassword(this.state.input).then(ok => {
      if (!ok) {
        // The input was not valid
        this.setState({ input: "" });
      }
    });
  }

  render() {
    return (
      <div style={Styles.popup}>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            justifyContent: "center",
            flex: 3
          }}
        >
          <h1>Login</h1>
        </div>
        <form
          style={{
            flex: 2,
            display: "flex",
            flexDirection: "column",
            width: "100%",
            justifyContent: "space-evenly"
          }}
          onSubmit={event => {
            event.preventDefault();
            this.handleInput();
          }}
        >
          <TextField
            label="Password or private key"
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
          <Button
            variant="contained"
            color="default"
            onClick={() => {
              this.popup.newKeyPair(KeyPair.fromRandom());
            }}
          >
            Create a new account
          </Button>
        </form>
      </div>
    );
  }
}
