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
  if (!fs.existsSync(filename)) {
    throw new Error(filename + " does not exist");
  }
  let stat = fs.lstatSync(filename);
  if (!stat.isFile()) {
    console.log("XXX stat:", stat);
    console.log("XXX stat.isFile() ", stat.isFile());
    console.log("XXX stat.isDirectory() ", stat.isDirectory());
    console.log("XXX stat.isBlockDevice() ", stat.isBlockDevice());
    console.log("XXX stat.isSymbolicLink() ", stat.isSymbolicLink());
    console.log("XXX stat.isCharacterDevice() ", stat.isCharacterDevice());
    console.log("XXX stat.isFIFO() ", stat.isFIFO());
    console.log("XXX stat.isSocket() ", stat.isSocket());
    throw new Error(filename + " isFile is false");
  }
  if (!isFile(filename)) {
    throw new Error(filename + " is not a file");
  }
  let serialized = fs.readFileSync(filename, "utf8");
  return KeyPair.fromSerialized(serialized);
}
