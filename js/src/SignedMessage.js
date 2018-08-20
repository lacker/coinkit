import KeyPair from "./KeyPair";

export default class SignedMessage {
  // Creates a signed message.
  // Users should generally not use this directly; use fromSigning or fromSerialized.
  // signer and signature are base64-encoded.
  // message is a JSONable object.
  constructor({ message, messageString, signer, signature }) {
    this.message = message;
    this.messageString = messageString;
    this.signer = signer;
    this.signature = signature;
  }

  // Construct a SignedMessage by signing a message.
  static fromSigning(message, keyPair) {
    if (!message) {
      throw new Error("cannot sign a falsy message");
    }
    let messageString = JSON.stringify(message);
    return new SignedMessage({
      message,
      messageString,
      signer: keyPair.getPublicKey(),
      signature: keyPair.sign(messageString)
    });
  }

  serialize() {
    return "e:" + this.signer + ":" + this.signature + ":" + this.messageString;
  }

  // Construct a SignedMessage from a serialized form
  // Throws an error if it receives an invalid message
  static fromSerialized(serialized) {
    let parts = serialized.split(":");
    if (parts.length < 4) {
      throw new Error("could not find 4 parts");
    }
    let [version, signer, signature] = parts.slice(0, 3);
    let messageString = parts.slice(3).join(":");
    if (version != "e") {
      throw new Error("unrecognized version");
    }
    if (!KeyPair.verifySignature(signer, messageString, signature)) {
      throw new Error("signature failed verification");
    }
    let message = JSON.parse(messageString);
    return new SignedMessage({ message, messageString, signer, signature });
  }
}
