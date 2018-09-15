// The root to display in the extension popup.

import React, { Component } from "react";

import Button from "@material-ui/core/Button";

import Client from "./Client";
import KeyPair from "./KeyPair";
import Login from "./Login";
import NewPassword from "./NewPassword";
import Status from "./Status";

export default class Popup extends Component {
  constructor(props) {
    super(props);

    this.state = {
      message: "hello world",
      keyPair: null,
      password: null,
      balance: null
    };
    this.client = new Client();

    this.newKeyPair = this.newKeyPair.bind(this);

    this.storage = chrome.extension.getBackgroundPage().storage;
    if (!this.storage) {
      throw new Error("cannot find storage");
    }
  }

  newKeyPair(kp) {
    this.client = new Client(kp);
    this.setState({
      keyPair: kp,
      password: null,
      balance: null
    });

    // Load the balance
    this.client.balance().then(balance => {
      if (this.state.keyPair == kp) {
        this.setState({
          balance: balance
        });
      }
    });
  }

  // Sets a new password for the already-existent keypair
  newPassword(password) {
    if (!this.state.keyPair) {
      throw new Error("cannot set new password with no keypair");
    }
    let data = {
      keyPair: this.state.keyPair.serialize()
    };

    this.storage.setPasswordAndData(password, data).then(() => {
      this.setState({
        password: password
      });
    });
  }

  // Tries to load stored data given the password that protects it
  // Returns whether the password was valid
  async checkPassword(password) {
    // TODO
    return false;
  }

  render() {
    let style = {
      display: "flex",
      alignSelf: "stretch",
      flexDirection: "column",
      justifyContent: "center"
    };
    if (!this.state.keyPair) {
      // Show the login screen
      return (
        <div style={style}>
          <Login popup={this} />
        </div>
      );
    }
    if (!this.state.password) {
      // They have a keypair but need to create a password.
      // Show the new-password screen
      return (
        <div style={style}>
          <NewPassword popup={this} />
        </div>
      );
    }

    // We have permissions for an account, so show its status
    return (
      <div style={style}>
        <Status
          popup={this}
          keyPair={this.state.keyPair}
          balance={this.state.balance}
        />
      </div>
    );
  }
}
