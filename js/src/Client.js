import Message from "./Message";

// Client is designed to be included in applications and run in untrusted application
// environment. It gets permissions by requesting them from the extension, whose code
// is trusted.
export default class Client {
  constructor() {
    // publicKey is null before permissions are acquired
    this.publicKey = null;

    // Callbacks are keyed by message id
    this.callbacks = {};

    window.addEventListener("message", event => {
      if (event.source != window || event.data.type != "fromCoinkit") {
        return;
      }

      let callback = this.callbacks[event.data.id];
      delete this.callbacks[event.data.id];
      if (!callback) {
        return;
      }

      callback(Message.fromSerialized(event.data.message));
    });
  }

  getMessageId() {
    return "" + Math.random();
  }

  async sendMessage(message) {
    let id = this.getMessageId();
    this.nextId++;
    let data = {
      id: id,
      type: "toCoinkit",
      message: message.serialize()
    };
    return new Promise((resolve, reject) => {
      this.callbacks[id] = resolve;
      window.postMessage(data);
    });
  }

  // Requests public key permission from the extension if we don't already have it.
  // Returns null if permission is denied.
  async getPublicKey() {
    // TODO
  }

  // Sends a query message, given the query properties.
  // Returns a promise for a message - a data message if the query worked, an error
  // message if it did not.
  async query(properties) {
    // TODO
  }
}
