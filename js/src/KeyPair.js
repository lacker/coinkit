// An ed25519 keypair. Designed to be parallel to the Go implementation.
// Annoyingly, our crypto library calls it a "secret key" while the Go library calls it
// a "private key". We try to name things "private key" when possible here.

import { toByteArray } from "base64-js";
import nacl from "tweetnacl";

// Adds padding to a base64-encoded string, which our library requires but some do not
function base64pad(s) {
  while (s.length % 4 != 0) {
    s += "=";
  }
  return s;
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
    let bytes = toByteArray(base64pad(priv));
    let keys = nacl.sign.keyPair.fromSecretKey(bytes);
    return new KeyPair(keys.publicKey, keys.secretKey);
  }
}
