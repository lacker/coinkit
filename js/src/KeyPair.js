// An ed25519 keypair. Designed to be parallel to the Go implementation.
// Annoyingly, our crypto library calls it a "secret key" while the Go library calls it
// a "private key". We try to name things "private key" when possible here.

import { fromByteArray, toByteArray } from "base64-js";
import nacl from "tweetnacl";
import forge from "node-forge";

// Adds padding to a base64-encoded string, which our library requires but some do not.
function bytesFromBase64(s) {
  while (s.length % 4 != 0) {
    s += "=";
  }
  return toByteArray(s);
}

// Decodes a Uint8array from a hex string.
function hexDecode(s) {
  if (s.length % 2 != 0) {
    throw new Error("hex-encoded byte arrays should be even length");
  }
  let length = s.length / 2;
  let answer = new Uint8Array(length);
  for (let i = 0; i < length; i++) {
    let chunk = s.substring(2 * i, 2 * i + 2);
    let value = parseInt(chunk, 16);
    if (value >= 256) {
      throw new Error(
        "bad byte value " + digit + " while decoding " + chunk + " from " + s
      );
    }
    answer[i] = value;
  }
  return answer;
}

export default class KeyPair {
  constructor(publicKey, privateKey) {
    this.publicKey = publicKey;
    this.privateKey = privateKey;

    if (publicKey.length != 32) {
      throw new Error(
        "public key length is " + publicKey.length + " but we expected 32"
      );
    }
    if (privateKey.length != 64) {
      throw new Error(
        "private key length is " + privateKey.length + " but we expected 64"
      );
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
    let pub = KeyPair.readPublicKey(j.Public);
    let priv = bytesFromBase64(j.Private);
    return new KeyPair(pub, priv);
  }

  // Returns the signature as base 64
  sign(message) {
    let sig = nacl.sign.detached(message, this.privateKey);
    return fromByteArray(sig);
  }

  // readPublicKey reads a public key from a string format.
  // This is parallel to Go's ReadPublicKey.
  // The string format starts with "0x" and is hex-encoded.
  // Throws an error if the input format is not valid.
  static readPublicKey(input) {
    if (input.length != 70) {
      throw new Error("public key " + input + " should be 70 characters long");
    }

    if (input.substring(0, 2) != "0x") {
      throw new Error("public key " + input + " should start with 0x");
    }

    // Check the checksum
    let key = hexDecode(input.substring(2, 66));
    let checksum1 = input.substring(66, 70);
    var md = forge.md.sha512.sha256.create();
    md.update(key);
    let checksum2 = md
      .digest()
      .toHex()
      .substring(0, 4);
    if (checksum1 != checksum2) {
      throw new Error(
        "mismatched checksums: " + checksum1 + " vs " + checksum2
      );
    }
  }
}
