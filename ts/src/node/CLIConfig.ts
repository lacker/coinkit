import * as fs from "fs";
import * as os from "os";
import * as path from "path";

const stringify = require("json-stable-stringify");

import { isDirectory } from "./FileUtil";

// A singleton object that keeps itself synced to disk.
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

  get(key) {
    return this.data[key];
  }

  set(key, value) {
    this.data[key] = value;
    this.write();
  }

  write() {
    fs.writeFileSync(this.filename, stringify(this.data));
  }
}
