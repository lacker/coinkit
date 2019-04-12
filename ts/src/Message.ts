const stringify = require("json-stable-stringify");

// Used to communicate with the blockchain
class Message {
  constructor(type, properties = {}) {
    if (typeof type !== "string") {
      throw new Error("Message must be constructed with a string type");
    }

    this.type = type;
    this._serialized = stringify({
      type,
      message: properties
    });
    for (let key in properties) {
      this[key] = properties[key];
    }
  }

  serialize() {
    return this._serialized;
  }

  static fromSerialized(serialized) {
    let { type, message } = JSON.parse(serialized);
    return new Message(type, message);
  }
}

module.exports = Message;
