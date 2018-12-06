// The root to display in the sample app.

import React, { Component } from "react";

import Client from "./Client";
import KeyPair from "./KeyPair";

import WebTorrent from "webtorrent";

export default class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      publicKey: null,
      mintBalance: null
    };

    this.client = new Client();
  }

  fetchBlockchainData() {
    // this.fetchBalance();
    this.fetchPublicKey();
  }

  fetchPeerData() {
    let client = new WebTorrent();

    let torrentId =
      "magnet:?xt=urn:btih:786005acb1312a764e90317e6b26ca3447583d12&dn=hello.txt&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com";

    client.add(torrentId, torrent => {
      torrent.on("done", () => {
        let file = torrent.files[0];
        console.log("got " + file.downloaded + " bytes");
        file.getBlob((err, blob) => {
          let reader = new FileReader();
          reader.onload = () => {
            console.log("result:", reader.result);
            alert(reader.result);
          };
          reader.readAsText(blob);
        });
      });
    });
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
            this.fetchBlockchainData();
          }}
        >
          Fetch Blockchain Data
        </button>
        <hr />
        <button
          onClick={() => {
            this.fetchPeerData();
          }}
        >
          Fetch Peer Data
        </button>
        <hr />
        <a href="http://hello.coinkit">Hello</a>
      </div>
    );
  }
}
