// A screen to show the status of your account.

import React, { Component } from "react";

import Styles from "./Styles";

export default class Status extends Component {
  // props.popup is a reference to the root popup
  // props.keyPair is the key pair
  // props.balance is the account balance, or null if unknown
  constructor(props) {
    super(props);

    this.popup = props.popup;
    this.keyPair = props.keyPair;
    this.balance = props.balance;
    this.logOut = this.logOut.bind(this);
  }

  logOut() {
    this.popup.newKeyPair(null);
  }

  render() {
    return (
      <div style={Styles.popup}>
        <h1>Status</h1>
        <div>Public key: {this.keyPair.publicKey}</div>
        <div>Balance: {this.balance == null ? "..." : this.balance}</div>
        <div style={{ color: "blue", cursor: "pointer" }} onClick={this.logOut}>
          Log out
        </div>
      </div>
    );
  }
}
