// This is the entry point for the hosting server.

// TODO: make this read args and create a HostingServer

const http = require("http");
const path = require("path");
const WebTorrent = require("webtorrent-hybrid");

const BlackHoleProxy = require("./BlackHoleProxy.js");
let proxy = new BlackHoleProxy(3333);

// TODO: remove this webtorrent seeding once the deploy-based seeding is working

// Seed a WebTorrent
let client = new WebTorrent();
let dir = path.resolve(__dirname, "samplesite");
client.seed(dir, torrent => {
  console.log("info hash: " + torrent.infoHash);

  torrent.on("wire", (wire, addr) => {
    console.log("connected to peer with address", addr);
  });
});
