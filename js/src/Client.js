import KeyPair from "./KeyPair";
import SignedMessage from "./SignedMessage";

// A client that handles interaction with the blockchain nodes.
export default class Client {
  // Sends a query message.
  // Returns a promise for a data message.
  async query(message) {
    let kp = KeyPair.fromRandom();
    let sm = SignedMessage.fromSigning(message, kp);
    let url = "http://localhost:8000/messages";
    let response = await fetch(url, {
      method: "POST",
      contentType: "text/plain; charset=utf-8",
      body: sm.serialize()
    });
    return response.json();
  }
}
