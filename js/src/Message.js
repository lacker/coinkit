export default class Message {
  constructor(type, properties) {
    this.type = type;
    this._serialized = JSON.stringify({
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
