// This is the entry point for the hosting server.

// TODO: make this read args for provider id and directory to store stuff in.
// TODO: create a HostingServer

const fs = require("fs");
const http = require("http");
const path = require("path");

const BlackHoleProxy = require("./BlackHoleProxy.js");
const HostingServer = require("./HostingServer.js");
const Tracker = require("./Tracker.js");

if (process.argv.length != 4) {
  console.log("usage: npm run host <providerID> <directory for hosting files>");
  process.exit(1);
}

let [node, hostmainjs, sid, directory] = process.argv;
let id = parseInt(sid);
if (id == 0) {
  console.log("bad id:", sid);
  process.exit(1);
}
if (!fs.existsSync(directory) || !fs.lstatSync(directory).isDirectory()) {
  console.log(directory, "is not a directory");
  process.exit(1);
}
console.log("hosting files for provider", id, "in", directory);
let host = new HostingServer(id, directory, true);

// Run a black hole proxy
let proxy = new BlackHoleProxy(3333);

// Run a tracker
let tracker = new Tracker(4444);
