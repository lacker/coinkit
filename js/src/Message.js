export default class Message {
  constructor(type, properties) {
    this.type = type;
    this._serialized = JSON.stringify({
      type,
      message: properties
    });
    Object.assign(properties, this);
  }

  serialize() {
    return this._serialized;
  }

  static fromSerialized(serialized) {
    let { type, message } = JSON.parse(serialized);
    return new Message(type, message);
  }
}
