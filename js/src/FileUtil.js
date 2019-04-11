// Utility functions that involve the filesystem and only work in Node

const fs = require("fs");

const KeyPair = require("./KeyPair.js");

function isDirectory(dir) {
  return fs.existsSync(dir) && fs.lstatSync(dir).isDirectory();
}

function isFile(filename) {
  return fs.existsSync(filename) && fs.lstatSync(filename).isFile();
}

function loadKeyPair(filename) {
  if (!isFile(filename)) {
    throw new Error(filename + " is not a file");
  }
  let serialized = fs.readFileSync(filename, "utf8");
  return KeyPair.fromSerialized(serialized);
}

module.exports = {
  isDirectory,
  isFile,
  loadKeyPair
};
