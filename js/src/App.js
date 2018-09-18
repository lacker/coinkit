// The root to display in the sample app.

import React, { Component } from "react";

import Client from "./Client";
import KeyPair from "./KeyPair";

export default class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      keyPair: null,
      mintBalance: null
    };

    this.client = new Client();
    this.fetchMintBalance();
  }

  async fetchMintBalance() {
    let mint =
      "0x32652ebe42a8d56314b8b11abf51c01916a238920c1f16db597ee87374515f4609d3";
    let query = {
      Account: mint
    };

    let balance = await this.client.query(query);
    this.setState({
      mintBalance: JSON.stringify(balance)
    });
  }

  login(privateKey) {
    // TODO: validate
    this.setState({ keyPair: KeyPair.fromPrivateKey(privateKey) });
  }

  render() {
    return (
      <div>
        <h1>this is the sample app</h1>
        <h1>
          {this.state.keyPair ? this.state.keyPair.publicKey : "nobody"} is
          logged in
        </h1>
        <h1>
          mint balance is{" "}
          {this.state.mintBalance == null ? "unknown" : this.state.mintBalance}
        </h1>
      </div>
    );
  }
}
