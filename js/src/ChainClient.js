// A client for talking to the blockchain servers.
// This code should work in both Node and in the browser.

// TODO: load this in some way that distinguishes between local testing, and prod
let URLS = [
  "http://localhost:8000",
  "http://localhost:8001",
  "http://localhost:8002",
  "http://localhost:8003"
];

export default class ChainClient {
  constructor() {
    this.listening = false;
  }

  listen() {
    this.listening = true;
    this.tick();
  }

  tick() {
    if (!this.listening) {
      return;
    }

    // TODO: update whatever data we're listening to

    setTimeout(() => this.tick(), 1000);
  }

  stopListening() {
    this.listening = false;
  }
}
