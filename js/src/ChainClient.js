const axios = require("axios");

const KeyPair = require("./KeyPair.js");
const Message = require("./Message.js");
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

function isEmpty(object) {
  for (let key in object) {
    return false;
  }
  return true;
}

class ChainClient {
  constructor(kp) {
    if (!kp) {
      kp = KeyPair.fromRandom();
    }
    this.keyPair = kp;
  }

  // Fetches the provider with the given provider id.
  // If there is no such provider, returns null.
  // If the server returns an error, we throw it.
  async getProvider(providerID) {
    let qm = new Message("Query", { providers: { id: providerID } });
    let dm = await this.sendMessage(qm);
    if (dm.type === "Error") {
      throw new Error(dm.error);
    }
    if (!dm.providers || !dm.providers[providerID]) {
      return null;
    }
    return dm.providers[providerID];
  }

  // Sends a Message upstream, signing with our keypair.
  // Returns a promise for the response Message.
  async sendMessage(message) {
    let clientMessage = SignedMessage.fromSigning(message, this.keyPair);
    let url = getServerURL() + "/messages";
    let body = clientMessage.serialize() + "\n";
    let response = await axios.post(url, body, {
      headers: { "Content-Type": "text/plain" },
      responseType: "text"
    });
    let text = response.data;
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
