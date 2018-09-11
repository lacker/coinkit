// The root to display in the extension popup.

import React, { Component } from "react";

import Button from "@material-ui/core/Button";

import Client from "./Client";
import KeyPair from "./KeyPair";
import Login from "./Login";
import NewPassword from "./NewPassword";

export default class Popup extends Component {
  constructor(props) {
    super(props);

    this.state = {
      message: "hello world",
      keyPair: null,
      password: null
    };
    this.client = new Client();

    this.newKeyPair = this.newKeyPair.bind(this);
    this.click = this.click.bind(this);
  }

  newKeyPair(kp) {
    this.setState({
      keyPair: kp,
      password: null
    });
  }

  // TODO: scrap this
  async click() {
    let mint =
      "0x32652ebe42a8d56314b8b11abf51c01916a238920c1f16db597ee87374515f4609d3";
    let query = {
      Account: mint
    };
    let response = await this.client.query(query);

    if (response.type == "Error") {
      this.setState({ message: "error: " + response.error });
    } else {
      this.setState({
        message: "balance is " + response.accounts[mint].balance
      });
    }
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

    // TODO: scrap this
    return (
      <div style={style}>
        <Button onClick={this.click}>load mint balance</Button>
        <h1>message: {this.state.message}</h1>
      </div>
    );
  }
}
