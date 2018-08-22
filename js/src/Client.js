import KeyPair from "./KeyPair";
import SignedMessage from "./SignedMessage";

// A client that handles interaction with the blockchain nodes.
export default class Client {
  // Sends a query message.
  // Returns a promise for a message - a data message if the query worked, an error
  // message if it did not.
  async query(message) {
    let m = {
      Type: "Query",
      Message: message
    };
    let kp = KeyPair.fromRandom();
    let sm = SignedMessage.fromSigning(m, kp);
    let url = "http://localhost:8000/messages";
    let body = sm.serialize() + "\n";
    let response = await fetch(url, {
      method: "post",
      body: body
    });
    let text = await response.text();
    console.log("XXX text:", text);
    return { error: "XXX TODO" };
  }
}
