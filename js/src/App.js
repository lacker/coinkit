// The root to display in the sample app.

import React, { Component } from "react";

import KeyPair from "./KeyPair";

export default class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      keyPair: null
    };
  }

  login(privateKey) {
    // TODO: validate
    this.setState({ keyPair: KeyPair.fromPrivateKey(privateKey) });
  }

  render(props) {
    return (
      <div>
        <h1>this is the sample app</h1>
        <h1>
          {this.state.keyPair ? this.state.keyPair.publicKey : "nobody"} is
          logged in
        </h1>
      </div>
    );
  }
}
