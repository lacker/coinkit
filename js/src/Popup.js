// The root to display in the extension popup.

import React, { Component } from "react";

import Button from "@material-ui/core/Button";

import Client from "./Client";
import KeyPair from "./KeyPair";
import Login from "./Login";

export default class Popup extends Component {
  constructor(props) {
    super(props);

    this.state = {
      message: "hello world",
      keyPair: null
    };
    this.client = new Client();

    this.setKeyPair = this.setKeyPair.bind(this);
    this.click = this.click.bind(this);
  }

  setKeyPair(kp) {
    this.setState({ keyPair: kp });
  }

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
      flexDirection: "column",
      justifyContent: "flex-start",
      alignItems: "center",
      width: 360,
      padding: 30
    };
    if (!this.state.keyPair) {
      // Show the login screen
      return (
        <div style={style}>
          <Login callback={this.setKeyPair} />
        </div>
      );
    }
    return (
      <div style={style}>
        <Button onClick={this.click}>load mint balance</Button>
        <h1>message: {this.state.message}</h1>
      </div>
    );
  }
}
