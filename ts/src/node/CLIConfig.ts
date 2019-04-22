import * as fs from "fs";
import * as os from "os";
import * as path from "path";

const stringify = require("json-stable-stringify");

import { isDirectory } from "./FileUtil";
import KeyPair from "../iso/KeyPair";

// An object that keeps itself synced to disk.
// The serialized form may contain these fields:
// keyPair: a plain-object keypair
export default class CLIConfig {
  constructor() {
    // If the directory doesn't exist, create it
    let dir = path.join(os.homedir(), ".coinkit");
    if (!isDirectory(dir)) {
      fs.mkdirSync(dir);
    }

    this.filename = path.join(dir, "config.json");

    if (isFile(this.filename)) {
      // If the config file exists, read it
      let str = fs.readFileSync(this.filename, "utf8");
      this.data = JSON.parse(str);
    } else {
      // If the config file does not exist, create it
      this.data = {};
      this.write();
    }
  }

  getKeyPair() {
    return KeyPair.fromPlain(this.data.keyPair);
  }

  setKeyPair(kp) {
    this.data.keyPair = kp.plain();
    this.write();
  }

  write() {
    fs.writeFileSync(this.filename, stringify(this.data));
  }
}
