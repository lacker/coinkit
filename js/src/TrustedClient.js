import KeyPair from "./KeyPair";
import Message from "./Message";
import SignedMessage from "./SignedMessage";

// A trusted client that handles interaction with the blockchain nodes.
// This client is trusted in the sense that it holds the user's keypair.
// This object is therefore only kept by the extension.

export default class TrustedClient {
  // Create a new client with no keypair.
  constructor() {
    this.keyPair = null;

    chrome.runtime.onMessage.addListener(
      (serializedMessage, sender, sendResponse) => {
        if (!sender.tab) {
          console.log("unexpected message from no tab:", serializedMessage);
          return false;
        }

        let message = Message.fromSerialized(serializedMessage);

        this.handleUntrustedMessage(message, sender.tab.url).then(
          responseMessage => {
            sendResponse(responseMessage.serialize());
          }
        );

        return true;
      }
    );
  }

  // Call from the background page
  static init() {
    window.client = new TrustedClient();
  }

  // Get the global trusted client from the background page
  static get() {
    let client = chrome.extension.getBackgroundPage().client;
    if (!client) {
      throw new Error("cannot find client");
    }
    return client;
  }

  sign(message) {
    let kp = this.keyPair || KeyPair.fromRandom();
    return SignedMessage.fromSigning(message, kp);
  }

  // Handles a message from an untrusted client.
  // Returns the message they should get back, or null if there is none.
  async handleUntrustedMessage(message, url) {
    // TODO: check permissions

    switch (message.type) {
      case "RequestPermission":
        return null;

      case "Query":
        let response = await this.sendMessage(message);
        return response;

      default:
        console.log("unexpected message type:", message.type);
        return null;
    }
  }

  // Sends a Message upstream, signing with our keypair.
  // Returns a promise for the response Message.
  async sendMessage(message) {
    let clientMessage = this.sign(message);
    let url = "http://localhost:8000/messages";
    let body = clientMessage.serialize() + "\n";
    let response = await fetch(url, {
      method: "post",
      body: body
    });
    let text = await response.text();
    let serialized = text.replace(/\n$/, "");

    // When there is an empty keepalive message from the server, we just return null
    let signed = SignedMessage.fromSerialized(serialized);
    if (signed == null) {
      return signed;
    }
    return signed.message;
  }

  // Sends a query message, given the query properties.
  // Returns a promise for a message - a data message if the query worked, an error
  // message if it did not.
  async query(properties) {
    let message = new Message("Query", properties);
    return this.sendMessage(message);
  }

  // Fetches the balance for this account
  async balance() {
    if (!this.keyPair) {
      return 0;
    }
    let pk = this.keyPair.getPublicKey();
    let query = {
      account: pk
    };
    let response = await this.query(query);
    let account = response.accounts[pk];
    if (!account) {
      return 0;
    }
    return account.balance;
  }
}
