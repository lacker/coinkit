// A screen to show the status of your account.

import React, { Component } from "react";

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
