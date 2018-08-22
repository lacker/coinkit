import KeyPair from "./KeyPair";
import SignedMessage from "./SignedMessage";

// A client that handles interaction with the blockchain nodes.
export default class Client {
  // Sends a query message.
  // Returns a promise for a signed message.
  async query(message) {
    let m = {
      Type: "Query",
      Message: message
    };
    let kp = KeyPair.fromRandom();
    let sm = SignedMessage.fromSigning(m, kp);
    let url = "http://localhost:8000/messages";
    let body = sm.serialize() + "\n";
    console.log("XXX body:", body);
    let response = await fetch(url, {
      method: "post",
      body: body
    });
    return response.text();
  }
}
