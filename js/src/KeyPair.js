// An ed25519 keypair. Designed to be parallel to the Go implementation.
// Annoyingly, our crypto library calls it a "secret key" while the Go library calls it
// a "private key". We try to name things "private key" when possible here.

import { fromByteArray, toByteArray } from "base64-js";
import nacl from "tweetnacl";

// Adds padding to a base64-encoded string, which our library requires but some do not.
function bytesFromBase64(s) {
  while (s.length % 4 != 0) {
    s += "=";
  }
  return toByteArray(s);
}

export default class KeyPair {
  constructor(publicKey, privateKey) {
    this.publicKey = publicKey;
    this.privateKey = privateKey;

    if (publicKey.length != 32 || privateKey.length != 64) {
      throw new Error("bad keys");
    }
  }

  static fromPrivateKey(priv) {
    let bytes = bytesFromBase64(priv);
    let keys = nacl.sign.keyPair.fromSecretKey(bytes);
    return new KeyPair(keys.publicKey, keys.secretKey);
  }

  // The input format is a serialized JSON string with 'Public' and 'Private' keys
  static fromSerialized(s) {
    let j = JSON.parse(s);
    let pub = bytesFromBase64(j.Public);
    let priv = bytesFromBase64(j.Private);
    return new KeyPair(pub, priv);
  }

  // Returns the signature as base 64
  sign(message) {
    let sig = nacl.sign.detached(message, this.privateKey);
    return fromByteArray(sig);
  }
}
