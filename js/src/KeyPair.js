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
  return Array.from(bytes)
    .map(byte => byte.toString(16).padStart(2, "0"))
    .join("");
}

// Creates a forge sha512/256 hash object from bytes
function forgeHash(bytes) {
  // Convert bytes to the format for bytes that forge wants
  let s = "";
  for (let i = 0; i < bytes.length; i++) {
    s += String.fromCharCode(bytes[i]);
  }
  let hash = forge.md.sha512.sha256.create();
  hash.update(s);
  return hash;
}

// Returns a hex checksum from a Uint8array public key.
function hexChecksum(bytes) {
  let hash = forgeHash(bytes);
  let digest = hash.digest();
  return digest.toHex().substring(0, 4);
}

// Returns a Uint8Array sha512_256 hash from a Uint8Array input.
function sha512_256(inputBytes) {
  let hash = forgeHash(inputBytes);
  let byteString = hash.digest().bytes();
  let outputBytes = new Uint8Array(32);
  for (let i = 0; i < 32; i++) {
    outputBytes[i] = byteString.charCodeAt(i);
  }
  return outputBytes;
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

  // Throws an error if priv is not a valid private key.
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
    let pub = KeyPair.decodePublicKey(j.Public);
    let priv = base64Decode(j.Private);
    return new KeyPair(pub, priv);
  }

  // Generates a keypair randomly
  static fromRandom() {
    let keys = nacl.sign.keyPair();
    return new KeyPair(keys.publicKey, keys.secretKey);
  }

  // Generates a keypair from a secret phrase
  static fromSecretPhrase(phrase) {
    // Hash the phrase for the ed25519 entropy seed bytes
    let bytes = new TextEncoder("utf-8").encode(phrase);
    let seed = sha512_256(bytes);
    let keys = nacl.sign.keyPair.fromSeed(seed);
    return new KeyPair(keys.publicKey, keys.secretKey);
  }

  // serialize() returns a serialized JSON string with 'Public' and 'Private' keys
  serialize() {
    let j = {
      Public: this.getPublicKey(),
      Private: base64Encode(this.privateKey)
    };

    // Pretty-encoding so that it matches our code style when saved to a file
    return JSON.stringify(j, null, 2) + "\n";
  }

  // We sign a string by utf-8 encoding it and signing the bytes.
  // Signatures are returned in base64 encoding.
  sign(string) {
    let bytes = new TextEncoder("utf-8").encode(string);
    let sig = nacl.sign.detached(bytes, this.privateKey);
    return base64Encode(sig);
  }

  // publicKey and signature are both base64-encoded strings
  // Returns whether the signature is legitimate.
  static verifySignature(publicKey, message, signature) {
    let key = KeyPair.decodePublicKey(publicKey);
    let msg = new TextEncoder("utf-8").encode(message);
    let sig = base64Decode(signature);
    try {
      return nacl.sign.detached.verify(msg, sig, key);
    } catch (e) {
      return false;
    }
  }

  // decodePublicKey reads a public key from a string format.
  // This is parallel to Go's ReadPublicKey.
  // The string format starts with "0x" and is hex-encoded.
  // Throws an error if the input format is not valid.
  static decodePublicKey(input) {
    if (input.length != 70) {
      throw new Error("public key " + input + " should be 70 characters long");
    }

    if (input.substring(0, 2) != "0x") {
      throw new Error("public key " + input + " should start with 0x");
    }

    // Check the checksum
    let key = hexDecode(input.substring(2, 66));
    let checksum1 = input.substring(66, 70);
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

  // Returns the public key in hex format
  getPublicKey() {
    return KeyPair.encodePublicKey(this.publicKey);
  }
}
