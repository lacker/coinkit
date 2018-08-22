// The root to display in the extension popup.

import React, { Component } from "react";

import Client from "./Client";

export default class Popup extends Component {
  constructor(props) {
    super(props);

    this.state = { message: "hello world" };
    this.client = new Client();

    this.click = this.click.bind(this);
  }

  async click() {
    let mint =
      "0x32652ebe42a8d56314b8b11abf51c01916a238920c1f16db597ee87374515f4609d3";
    let query = {
      Account: mint
    };
    let response = await this.client.query(query);

    if (response.error) {
      this.setState({ message: "error: " + response.error });
    } else {
      this.setState({ message: "balance is " + response.Accounts[mint] });
    }
  }

  render(props) {
    return (
      <div>
        <button onClick={this.click}>load mint balance</button>
        <h1>message: {this.state.message}</h1>
      </div>
    );
  }
}
