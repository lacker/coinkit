// An ed25519 keypair. Designed to be parallel to the Go implementation.
// Annoyingly, our crypto library calls it a "secret key" while the Go library calls it
// a "private key". We try to name things "private key" when possible here.

import { fromByteArray, toByteArray } from "base64-js";
import nacl from "tweetnacl";
import forge from "node-forge";
import { TextEncoder } from "text-encoding-shim";

// Decodes a Uint8Array from a base64 string.
// Adds = padding at the end, which our library requires but some do not.
function base64Decode(s) {
  while (s.length % 4 != 0) {
    s += "=";
  }
  return toByteArray(s);
}

// Encodes a Uint8array into a base64 string.
// Removes any = padding at the end.
function base64Encode(bytes) {
  let padded = fromByteArray(bytes);
  return padded.replace(/=*$/, "");
}

// Decodes a Uint8Array from a hex string.
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
        "bad byte value " + value + " while decoding " + chunk + " from " + s
      );
    }
    answer[i] = value;
  }
  return answer;
}

// Encodes a Uint8Array into a hex string.
function hexEncode(bytes) {
  return bytes.reduce((str, byte) => str + byte.toString(16).padStart(2, "0"));
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
    let bytes = base64Decode(priv);
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
    let priv = base64Decode(j.Private);
    return new KeyPair(pub, priv);
  }

  // serialize() returns a serialized JSON string with 'Public' and 'Private' keys
  serialize() {
    // XXX
  }

  // Returns the signature as base 64.
  // Strips equal signs for Go compatibility
  sign(bytes) {
    let sig = nacl.sign.detached(bytes, this.privateKey);
    let padded = fromByteArray(sig);
    return padded.replace(/=*$/, "");
  }

  // utf-8 encodes a string before signing
  signString(string) {
    let arr = new TextEncoder("utf-8").encode(string);
    return this.sign(arr);
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

  // encodePublicKey creates a string-format public key from Uint8Array.
  // The checksum is added at the end
  static encodePublicKey(key) {
    if (key.length != 32) {
      throw new Error("public keys should be 32 bytes long");
    }
    return "0x" + hexEncode(key) + hexChecksum(key);
  }
}
