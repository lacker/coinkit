// An ed25519 keypair. Designed to be parallel to the Go implementation.
// Annoyingly, our crypto library calls it a "secret key" while the Go library calls it
// a "private key". We try to name things "private key" when possible here.

import nacl from "tweetnacl";

export default class KeyPair {
  // TODO: replace placeholder
  constructor({ publicKey, privateKey }) {
    this.publicKey = publicKey;
    this.privateKey = privateKey;
  }

  static fromPrivateKey(priv) {
    let keys = nacl.sign.keyPair.fromSecretKey(priv);
    return new KeyPair({
      publicKey: keys.publicKey,
      privateKey: keys.secretKey
    });
  }
}
