// The root to display in the extension popup.

import React, { Component } from "react";
import { connect } from "react-redux";
import Button from "@material-ui/core/Button";

import KeyPair from "./KeyPair";
import Login from "./Login";
import NewPassword from "./NewPassword";
import RequestPermission from "./RequestPermission";
import Status from "./Status";
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

import {
  loadFromStorage,
  logOut,
  newKeyPair,
  newPassword,
  denyPermission,
  grantPermission
} from "./actions";

class Popup extends Component {
  constructor(props) {
    super(props);
  }

  logOut() {
    this.props.dispatch(logOut());
  }

  newKeyPair(kp) {
    this.props.dispatch(newKeyPair(kp));
  }

  // Sets a new password for the already-existent keypair
  newPassword(password) {
    this.props.dispatch(newPassword(password));
  }

  // Tries to load a stored keypair given the password that protects it.
  // Returns whether the password was valid
  async checkPassword(password) {
    let storage = await Storage.get();
    let ok = await storage.checkPassword(password);
    if (!ok) {
      console.log("bad password:", password);
      return false;
    }
    this.props.dispatch(loadFromStorage(storage));
    return true;
  }

  render() {
    let style = {
      display: "flex",
      alignSelf: "stretch",
      flexDirection: "column",
      justifyContent: "center"
    };
    if (!this.props.keyPair) {
      // Show the login screen
      return (
        <div style={style}>
          <Login popup={this} />
        </div>
      );
    }
    if (!this.props.password) {
      // They have a keypair but need to create a password.
      // Show the new-password screen
      return (
        <div style={style}>
          <NewPassword popup={this} />
        </div>
      );
    }

    if (this.props.request) {
      let host = this.props.request.host;
      let permissions = this.props.request.permissions;
      // The app is requesting permissions
      return (
        <div style={style}>
          <RequestPermission
            host={host}
            permissions={permissions}
            accept={() => {
              this.props.dispatch(grantPermission(host, permissions));
            }}
            deny={() => {
              this.props.dispatch(denyPermission());
            }}
          />
        </div>
      );
    }

    // We have permissions for an account, so show its status
    return (
      <div style={style}>
        <Status popup={this} keyPair={this.props.keyPair} />
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    password: state.password,
    keyPair: state.keyPair,
    request: state.request
  };
}

export default connect(mapStateToProps)(Popup);
