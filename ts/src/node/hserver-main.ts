// This is the entry point for the hosting server.

import * as http from "http";
import * as os from "os";
import * as path from "path";

import * as args from "args";

args
  .option("tracker", "The port on which the tracker will be running", 4000)
  .option("proxy", "The port on which the proxy will be running", 3000)
  .option("id", "The provider id to host files for", 0)
  .option("keypair", "File containing keys for the user to host files for", "")
  .option("capacity", "Amount of space in megabytes to host", 0)
  .option(
    "directory",
    "The directory to store files in",
    path.join(os.homedir(), "hostfiles")
  );

const flags = args.parse(process.argv);

import BlackHoleProxy from "./BlackHoleProxy";
import HostingServer from "./HostingServer";
import Tracker from "./Tracker";

if (flags.capacity <= 0) {
  console.log("to host files you must set --capacity");
  process.exit(1);
}

if (flags.id < 1 && flags.keypair.length < 1) {
  console.log(
    flags.id,
    "you must specify a provider with either --id or --keypair"
  );
  process.exit(1);
}

let options = {
  id: undefined,
  keyPair: undefined,
  capacity: flags.capacity,
  directory: flags.directory,
  verbose: true
};
if (flags.id >= 1) {
  options.id = flags.id;
} else {
  options.keyPair = flags.keypair;
}

let host = new HostingServer(options);
host.serve();

// Run a black hole proxy
let proxy = new BlackHoleProxy(flags.proxy);

// Run a tracker
let tracker = new Tracker(flags.tracker);
