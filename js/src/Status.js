// A screen to show the status of your account.

import React, { Component } from "react";
import Button from "@material-ui/core/Button";

import Styles from "./Styles";

export default class Status extends Component {
  // props.popup is a reference to the root popup
  // props.keyPair is the key pair
  // props.balance is the account balance, or null if unknown
  render() {
    if (this.props.balance == null) {
      this.props.popup.loadBalance();
    }

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
              {this.props.keyPair.getPublicKey()}
            </div>
          </div>
          <div>
            Balance:
            <div>
              {this.props.balance == null ? "loading..." : this.props.balance}
            </div>
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
              this.props.popup.logOut();
            }}
          >
            Log out
          </Button>
        </div>
      </div>
    );
  }
}
