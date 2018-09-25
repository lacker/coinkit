// The root to display in the extension popup.

import React, { Component } from "react";

import Button from "@material-ui/core/Button";

import KeyPair from "./KeyPair";
import Login from "./Login";
import NewPassword from "./NewPassword";
import Status from "./Status";
import TrustedClient from "./TrustedClient";

export default class Popup extends Component {
  constructor(props) {
    super(props);

    this.client = chrome.extension.getBackgroundPage().client;
    if (!this.client) {
      throw new Error("cannot find client");
    }

    this.storage = chrome.extension.getBackgroundPage().storage;
    if (!this.storage) {
      throw new Error("cannot find storage");
    }

    this.state = this.stateFromStorage();
  }

  // Updates the client keypair as a side effect
  // TODO: This isn't good design. Maybe the client should look directly at the storage
  stateFromStorage() {
    let clear = {
      keyPair: null,
      password: null,
      balance: null
    };

    if (!this.storage.data) {
      return clear;
    }

    if (typeof this.storage.data != "object") {
      console.log("bad stored data:", this.storage.data);
      return clear;
    }

    let kp;
    try {
      kp = KeyPair.fromSerialized(this.storage.data.keyPair);
    } catch (e) {
      console.log("invalid keypair from storage:", this.storage.data);
      return clear;
    }

    this.client.setKeyPair(kp);
    return {
      keyPair: kp,
      password: this.storage.password,
      balance: null
    };
  }

  logOut() {
    this.storage.logOut();
    this.newKeyPair(null);
  }

  newKeyPair(kp) {
    this.client.setKeyPair(kp);
    this.setState({
      keyPair: kp,
      password: null,
      balance: null
    });
  }

  loadBalance() {
    if (this.state.balance != null) {
      return;
    }
    let kp = this.state.keyPair;
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
    let data = {
      keyPair: this.state.keyPair.serialize()
    };
    this.storage.setPasswordAndData(password, data).then(() => {
      this.setState({
        password: password
      });
    });
  }

  // Tries to load a stored keypair given the password that protects it.
  // Returns whether the password was valid
  async checkPassword(password) {
    let ok = await this.storage.checkPassword(password);
    if (!ok) {
      console.log("bad password:", password);
      return false;
    }
    let state = this.stateFromStorage();
    this.setState(state);
    return true;
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
