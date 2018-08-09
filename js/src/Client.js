// A client that handles interaction with the blockchain nodes.

export default class Client {
  // Sends a query message.
  // Returns a promise for a data message.
  async query(message) {
    let url = "http://localhost:9090/api";
    let response = await fetch(url);
    return response.json();
  }
}
