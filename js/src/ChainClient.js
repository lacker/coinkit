const axios = require("axios");

const KeyPair = require("./KeyPair.js");
const SignedMessage = require("./SignedMessage.js");

// A client for talking to the blockchain servers.
// This client only uses one keypair across its lifetime.
// It is not expensive to set up, though, so if you have a different keypair just
// create a different client object.
// This code should work in both Node and in the browser.

// TODO: load this in some way that distinguishes between local testing, and prod
let URLS = [
  "http://localhost:8000",
  "http://localhost:8001",
  "http://localhost:8002",
  "http://localhost:8003"
];

function getServerURL() {
  let index = Math.floor(Math.random() * URLS.length);
  return URLS[index];
}

class ChainClient {
  constructor(kp) {
    if (!kp) {
      kp = KeyPair.fromRandom();
    }
    this.keyPair = kp;
    this.listening = false;
  }

  listen() {
    this.listening = true;
    this.tick();
  }

  tick() {
    if (!this.listening) {
      return;
    }

    // TODO: update whatever data we're listening to

    setTimeout(() => this.tick(), 1000);
  }

  stopListening() {
    this.listening = false;
  }

  // Sends a Message upstream, signing with our keypair.
  // Returns a promise for the response Message.
  async sendMessage(message) {
    let clientMessage = SignedMessage.fromSigning(message, this.keyPair);
    let url = getServerURL() + "/messages";
    let body = clientMessage.serialize() + "\n";
    let text = await axios.post(url, body, {
      headers: { "Content-Type": "text/plain" },
      responseType: "text"
    });
    let serialized = text.replace(/\n$/, "");

    // When there is an empty keepalive message from the server, we just return null
    let signed = SignedMessage.fromSerialized(serialized);
    if (signed == null) {
      return signed;
    }
    return signed.message;
  }
}

module.exports = ChainClient;
