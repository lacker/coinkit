import Message from "./Message";
import { missingPermissions, hasPermission } from "./permission.js";

// Client is designed to be included in applications and run in an untrusted application
// environment. It gets permissions by requesting them from the extension, whose code
// is trusted.
//
// There are two types of messages - browser messages are used to communicate between
// application page, content script, and other extension logic. The Message object is used
// to communicate with the blockchain.
//
// Browser messages are plain json that contains:
// id: a random id string specifying this message
// type: either "toCoinkit" or "fromCoinkit" for whether it is upstream or downstream
// message: a serialized blockchain message
//
// When the extension sends a response message, it includes the same id as the message
// that it was responding to.

export default class Client {
  constructor() {
    // publicKey is null before permissions are acquired
    this.publicKey = null;

    // Callbacks are keyed by message id
    this.callbacks = {};

    // We store the most recent permission message we received from the extension.
    // XXX: on startup, initialize this
    // Before we receive any permission message, we just assume we have no permissions.
    this.permissions = new Message("Permission");

    window.addEventListener("message", event => {
      if (
        event.source != window ||
        event.data.type != "fromCoinkit" ||
        !event.data.message
      ) {
        return;
      }

      let message = Message.fromSerialized(event.data.message);
      console.log("XXX got message from extension:", message);

      if (message.type == "Permission") {
        this.permissions = message;
      }

      let callback = this.callbacks[event.data.id];
      delete this.callbacks[event.data.id];
      if (!callback) {
        return;
      }

      callback(message);
    });
  }

  // Each browser message has an id
  getMessageId() {
    return "" + Math.random();
  }

  async sendMessage(message) {
    console.log("XXX sending message to extension:", message);
    let id = this.getMessageId();
    this.nextId++;
    let data = {
      id: id,
      type: "toCoinkit",
      message: message.serialize()
    };
    return new Promise((resolve, reject) => {
      this.callbacks[id] = resolve;
      window.postMessage(data, "*");
    });
  }

  // Returns a permission message containing all the permissions.
  // Throws an error if the user denies permission.
  async requestPermission(permissions) {
    throw new Error("XXX implement me");
  }

  // Requests public key permission from the extension if we don't already have it.
  // Throws an error if the user denies permission.
  async getPublicKey() {
    if (!hasPermission(this.permissions, { publicKey: true })) {
      await this.requestPermission({ publicKey: true });
    }
    let message = new Message("Query", { publicKey: true });
    let response = await this.sendMessage(message);
    this.publicKey = response.publicKey;
    return this.publicKey;
  }

  // Sends a query message, given the query properties.
  // Returns a promise for a message - a data message if the query worked, an error
  // message if it did not.
  async query(properties) {
    let message = new Message("Query", properties);
    return this.sendMessage(message);
  }
}
