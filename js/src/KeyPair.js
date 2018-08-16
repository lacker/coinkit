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

// Returns a hex checksum from a Uint8array public key.
function hexChecksum(bytes) {
  // Convert bytes to the format for bytes that forge wants
  let s = "";
  for (let i = 0; i < bytes.length; i++) {
    s += String.fromCharCode(bytes[i]);
  }
  let hash = forge.md.sha512.sha256.create();
  hash.update(s);
  let digest = hash.digest();
  return digest.toHex().substring(0, 4);
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
    if (!j.Public) {
      throw new Error("serialized key pair must have Public field");
    }
    if (!j.Private) {
      throw new Error("serialized key pair must have Private field");
    }
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
    let md = forge.md.sha512.sha256.create();
    md.update(key);
    let checksum2 = hexChecksum(key);
    if (checksum1 != checksum2) {
      throw new Error(
        "mismatched checksums: " + checksum1 + " vs " + checksum2
      );
    }

    return key;
  }

  // Testing that our JavaScript libraries work like our Go libraries
  static testCryptoBasics() {
    let hash = forge.md.sha512.sha256.create();
    let sum = hash.digest().getBytes();
    if (sum.charCodeAt(0) != 198) {
      throw new Error("first byte of hashed nothing should be 198");
    }

    hash = forge.md.sha512.sha256.create();
    hash.update("qq", "utf-8");
    sum = hash.digest().getBytes();
    expect(sum.charCodeAt(0)).toBe(59);

    let bytes =
      String.fromCharCode(1) +
      String.fromCharCode(2) +
      String.fromCharCode(3) +
      String.fromCharCode(4);
    hash = forge.md.sha512.sha256.create();
    hash.update(bytes);
    sum = hash.digest().getBytes();
    expect(sum.charCodeAt(0)).toBe(254);
  }
}
