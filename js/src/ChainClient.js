const axios = require("axios");

const KeyPair = require("./KeyPair.js");
const Message = require("./Message.js");
const SignedMessage = require("./SignedMessage.js");

// A client for talking to the blockchain servers.
// This client only uses one keypair across its lifetime.
// It is not expensive to set up, though, so if you have a different keypair just
// create a different client object.
// This code should work in both Node and in the browser.

// TODO: load this in some way that distinguishes between local testing, and prod
let URLS = [
  "http://localhost:8000",
  "http://localhost:8001",
  "http://localhost:8002",
  "http://localhost:8003"
];

const STANDARD_WAIT = 1000;

async function standardWait() {
  let promise = new Promise((resolve, reject) => {
    setTimeout(resolve, STANDARD_WAIT);
  });
  return await promise;
}

function getServerURL() {
  let index = Math.floor(Math.random() * URLS.length);
  return URLS[index];
}

function isEmpty(object) {
  for (let key in object) {
    return false;
  }
  return true;
}

class ChainClient {
  constructor(kp) {
    if (!kp) {
      kp = KeyPair.fromRandom();
    }
    this.keyPair = kp;
    this.verbose = false;
  }

  log(...args) {
    if (this.verbose) {
      console.log(...args);
    }
  }

  // Keeps re-fetching the provider until it exists.
  async waitForProvider(providerID) {
    this.log("waiting for provider", providerID, "to be created");
    while (true) {
      let provider = await this.getProvider(providerID);
      if (provider !== null) {
        return provider;
      }
      this.log("not yet, still waiting");
      await standardWait();
    }
  }

  // Fetches the provider with the given provider id.
  // If there is no such provider, returns null.
  async getProvider(providerID) {
    let dm = await this.query({ providers: { id: providerID } });
    if (!dm.providers || !dm.providers[providerID]) {
      return null;
    }
    return dm.providers[providerID];
  }

  // Returns the information for the newly-created provider.
  async createProvider(capacity) {
    // To figure out which provider is newly-created, we need to check existing ones
    let user = this.keyPair.getPublicKey();
    let initialProviders = await this.getProviders({ owner: user });
    this.log("some providers already exist:", Object.keys(initialProviders));
    await this.sendOperation("CreateProvider", { capacity });
    let providers = await this.getProviders({ owner: user });

    for (let provider of providers) {
      if (!initialProviders[provider.id]) {
        // This provider wasn't in the initial set
        return provider;
      }
    }

    throw new Error("no provider seems to have been created");
  }

  // Returns once the operation has been accepted into the blockchain.
  // Signer, fee, and sequence are all added.
  // Throws an error if there is no matching user account.
  // TODO: there is a race condition where a different operation with the same sequence
  // number could be sent. We should detect that.
  async sendOperation(type, operation) {
    // First check that we have an account
    let user = this.keyPair.getPublicKey();
    let account = await this.getAccount(user);
    if (!account) {
      throw new Error("cannot create provider for a nonexistent user account");
    }

    // Make a signed op without the actual signature
    let newSequence = account.sequence + 1;
    let sop = {
      type: type,
      operation: {
        ...operation,
        fee: 0,
        sequence: newSequence
      }
    };

    let opm = new Message("Operation", { operations: [sop] });
    let sopm = this.keyPair.signOperationMessage(opm);
    let response = await this.sendMessage(sopm);

    // TODO: check for responses that indicate trouble

    // Wait for the op to be processed
    while (true) {
      let account = await this.getAccount(user);
      if (account.sequence === newSequence) {
        break;
      }
      await standardWait();
    }
  }

  // Fetches data for the listed buckets.
  // Returns an object mapping bucket name to bucket data.
  async getBuckets(query) {
    let dm = await this.query({ buckets: query });
    let answer = {};
    for (let bucket of dm.buckets) {
      answer[bucket.name] = bucket;
    }
    return answer;
  }

  // Fetches data for providers according to the given query.
  // Returns an object mapping provider id to provider data.
  async getProviders(query) {
    let dm = await this.query({ providers: query });
    let answer = {};
    for (let provider of dm.providers) {
      answer[provider.id] = provider;
    }
    return answer;
  }

  // Fetches the account with the given user, or null if there is no such account.
  async getAccount(user) {
    let dm = await this.query({ account: user });
    if (!dm.accounts || !dm.accounts[user]) {
      return null;
    }
    return dm.accounts[user];
  }

  // Sends a query message. Returns the data message response.
  // Throws an error if we get an error message back.
  async query(params) {
    let qm = new Message("Query", params);
    let dm = await this.sendMessage(qm);
    if (dm.type === "Error") {
      throw new Error(dm.error);
    }
    return dm;
  }

  // Sends a Message upstream, signing with our keypair.
  // Returns a promise for the response Message.
  async sendMessage(message) {
    let clientMessage = SignedMessage.fromSigning(message, this.keyPair);
    let url = getServerURL() + "/messages";
    let body = clientMessage.serialize() + "\n";
    let response = await axios.post(url, body, {
      headers: { "Content-Type": "text/plain" },
      responseType: "text"
    });
    let text = response.data;
    let serialized = text.replace(/\n$/, "");

    // When there is an empty keepalive message from the server, we just return null
    let signed = SignedMessage.fromSerialized(serialized);
    if (signed == null) {
      return signed;
    }
    return signed.message;
  }
}

module.exports = ChainClient;
