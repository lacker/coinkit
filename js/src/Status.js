// A screen to show the status of your account.

import React, { Component } from "react";
import Button from "@material-ui/core/Button";

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
  }

  logOut() {
    this.popup.newKeyPair(null);
  }

  render() {
    return (
      <div style={Styles.popup}>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-evenly",
            width: "100%",
            flex: 3
          }}
        >
          <h1>Status</h1>
          <div>
            Public key:
            <div
              style={{
                wordWrap: "break-word"
              }}
            >
              {this.keyPair.getPublicKey()}
            </div>
          </div>
          <div>
            Balance:
            <div>{this.balance == null ? "loading..." : this.balance}</div>
          </div>
        </div>
        <div
          style={{
            flex: 2,
            display: "flex",
            flexDirection: "column",
            width: "100%",
            justifyContent: "space-evenly"
          }}
        >
          <Button
            variant="contained"
            color="default"
            onClick={() => {
              this.popup.newKeyPair(null);
            }}
          >
            Log out
          </Button>
        </div>
      </div>
    );
  }
}
