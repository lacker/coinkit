// The root to display in the extension popup.

import React, { Component } from "react";
import { connect } from "react-redux";
import Button from "@material-ui/core/Button";

import KeyPair from "./KeyPair";
import Login from "./Login";
import NewPassword from "./NewPassword";
import Status from "./Status";
import TrustedClient from "./TrustedClient";

class Popup extends Component {
  constructor(props) {
    super(props);

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
      password: null
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

    TrustedClient.get().setKeyPair(kp);
    return {
      keyPair: kp,
      password: this.storage.password
    };
  }

  logOut() {
    this.storage.logOut();
    this.newKeyPair(null);
  }

  newKeyPair(kp) {
    TrustedClient.get().setKeyPair(kp);
    this.setState({
      keyPair: kp,
      password: null
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
        <Status popup={this} keyPair={this.state.keyPair} />
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    password: state.password,
    keyPair: state.keyPair
  };
}

export default connect(mapStateToProps)(Popup);
