// This is the entry point for the hosting server.

// TODO: make this read args for provider id and directory to store stuff in.
// TODO: create a HostingServer

const http = require("http");
const path = require("path");

if (process.argv.length != 4) {
  console.log("usage: npm run host <providerID> <directory for hosting files>");
  process.exit(1);
}

let [_, _, id, directory] = process.argv;
console.log("id:", id);
console.log("directory:", directory);

// Run a black hole proxy
const BlackHoleProxy = require("./BlackHoleProxy.js");
let proxy = new BlackHoleProxy(3333);

// Run a tracker
const Tracker = require("./Tracker.js");
let tracker = new Tracker(4444);
