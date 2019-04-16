// Utility functions that involve the filesystem and only work in Node

import * as fs from "fs";

import KeyPair from "../iso/KeyPair";

export function isDirectory(dir) {
  return fs.existsSync(dir) && fs.lstatSync(dir).isDirectory();
}

export function isFile(filename) {
  return fs.existsSync(filename) && fs.lstatSync(filename).isFile();
}

export function loadKeyPair(filename) {
  if (!isFile(filename)) {
    throw new Error(filename + " is not a file");
  }
  let serialized = fs.readFileSync(filename, "utf8");
  return KeyPair.fromSerialized(serialized);
}
