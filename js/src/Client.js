import KeyPair from "./KeyPair";
import Message from "./Message";
import SignedMessage from "./SignedMessage";

// A client that handles interaction with the blockchain nodes.
export default class Client {
  // Create a new client with the provided keypair.
  // If no keypair is provided, use a random one.
  constructor(keyPair) {
    if (!keyPair) {
      this.keyPair = KeyPair.fromRandom();
    } else {
      this.keyPair = keyPair;
    }
  }

  // Sends a Message upstream, signing with our keypair.
  // Returns a promise for the response Message.
  async sendMessage(message) {
    let clientMessage = SignedMessage.fromSigning(message, this.keyPair);
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
    let queryMessage = new Message("Query", properties);
    return this.sendMessage(queryMessage);
  }

  // Fetches the balance for this account
  async balance() {
    let pk = this.keyPair.getPublicKey();
    let query = {
      account: pk
    };
    let response = await this.query(query);
    return response.accounts[pk].balance;
  }
}
