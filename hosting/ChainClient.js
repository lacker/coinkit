// A Node client for talking to the blockchain servers
// TODO: think about how to combine some of this code with the web js client

// TODO: load this in some way that distinguishes between local testing, and prod
let URLS = [
  "http://localhost:8000",
  "http://localhost:8001",
  "http://localhost:8002",
  "http://localhost:8003"
];

export default class ChainClient {
  constructor() {
    this.active = true;
    this.tick();
  }

  tick() {
    // TODO: do stuff
  }

  stop() {}
}
