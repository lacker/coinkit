// This is the entry point for the hosting server.

// TODO: make this read args and create a HostingServer

const http = require("http");
const path = require("path");

const WebTorrent = require("webtorrent-hybrid");

// Run a black hole proxy
const BlackHoleProxy = require("./BlackHoleProxy.js");
let proxy = new BlackHoleProxy(3333);

// Run a tracker
const Tracker = require("./Tracker.js");
let tracker = new Tracker(4444);
