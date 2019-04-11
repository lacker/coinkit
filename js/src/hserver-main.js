// This is the entry point for the hosting server.

const http = require("http");
const os = require("os");
const path = require("path");

const args = require("args");

args
  .option("tracker", "The port on which the tracker will be running", 4000)
  .option("proxy", "The port on which the proxy will be running", 3000)
  .option("id", "The provider id to host files for", 0)
  .option("keypair", "File containing keys for the user to host files for", "")
  .option(
    "directory",
    "The directory to store files in",
    path.join(os.homedir(), "hostfiles")
  );

const flags = args.parse(process.argv);

const BlackHoleProxy = require("./BlackHoleProxy.js");
const HostingServer = require("./HostingServer.js");
const Tracker = require("./Tracker.js");

if (flags.id < 1 && flags.keypair.length < 1) {
  console.log(
    flags.id,
    "you must specify a provider with either --id or --keypair"
  );
  process.exit(1);
}

console.log("hosting files for provider", flags.id, "in", flags.directory);
let host = new HostingServer(flags.id, flags.directory, true);
host.serve();

// Run a black hole proxy
let proxy = new BlackHoleProxy(flags.proxy);

// Run a tracker
let tracker = new Tracker(flags.tracker);
