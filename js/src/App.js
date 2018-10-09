// The root to display in the sample app.

import React, { Component } from "react";

import Client from "./Client";
import KeyPair from "./KeyPair";

export default class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      publicKey: null,
      mintBalance: null
    };

    this.client = new Client();
  }

  fetchData() {
    // this.fetchBalance();
    this.fetchPublicKey();
  }

  async fetchBalance() {
    let mint =
      "0x32652ebe42a8d56314b8b11abf51c01916a238920c1f16db597ee87374515f4609d3";
    let query = {
      account: mint
    };

    let response = await this.client.query(query);
    if (!response.accounts || !response.accounts[mint]) {
      console.log("bad message:", response);
    } else {
      let balance = response.accounts[mint].balance;
      this.setState({
        mintBalance: balance
      });
    }
  }

  async fetchPublicKey() {
    let pk = await this.client.getPublicKey();
    this.setState({
      publicKey: pk
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
          {this.state.publicKey ? this.state.publicKey : "nobody"} is logged in
        </h1>
        <h1>
          mint balance is{" "}
          {this.state.mintBalance == null ? "unknown" : this.state.mintBalance}
        </h1>
        <button
          onClick={() => {
            this.fetchData();
          }}
        >
          Fetch Data
        </button>
      </div>
    );
  }
}
