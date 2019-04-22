import * as os from "os";
import * as path from "path";

// A singleton object that keeps itself synced to disk.
export default class CLIConfig {
  constructor() {
    // If the directory doesn't exist, create it
    let dir = path.join(os.homedir(), ".coinkit");
  }
}
