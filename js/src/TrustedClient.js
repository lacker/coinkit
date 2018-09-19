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
        console.log("XXX message from", sender.tab);

        if (!sender.tab) {
          console.log("unexpected message from no tab:", serializedMessage);
          return false;
        }

        let message = Message.fromSerialized(serializedMessage);

        // TODO: handle a permissions request

        this.handleUntrustedMessage(message, sender.tab.url).then(
          responseMessage => {
            sendResponse(responseMessage.serialize());
          }
        );

        return true;
      }
    );
  }

  setKeyPair(kp) {
    this.keyPair = kp;
  }

  sign(message) {
    let kp = this.keyPair || KeyPair.fromRandom();
    return SignedMessage.fromSigning(message, kp);
  }

  async handleUntrustedMessage(message, url) {
    // TODO: check permissions
    return this.sendMessage(message);
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
    return response.accounts[pk].balance;
  }
}
