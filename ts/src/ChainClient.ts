const axios = require("axios");

const KeyPair = require("./KeyPair.js");
const Message = require("./Message.js");
const SignedMessage = require("./SignedMessage.js");
const { sleep } = require("./Util.js");

// A client for talking to the blockchain servers.
// This client only uses one keypair across its lifetime.
// It is not expensive to set up, though, so if you have a different keypair just
// create a different client object.
// This code should work in both Node and in the browser.

// TODO: load urls differently for local, testing, and prod at runtime

// For general local operation
let LOCAL_URLS = [
  "http://localhost:8000",
  "http://localhost:8001",
  "http://localhost:8002",
  "http://localhost:8003"
];

// For easy-to-debug operation, only hit a single server
let DEBUG_URLS = ["http://localhost:8000"];

let URLS = DEBUG_URLS;

const STANDARD_WAIT = 1000;

async function standardWait() {
  return await sleep(STANDARD_WAIT);
}

async function retryPost(url, body, params) {
  let retries = 3;
  for (let i = 0; i < retries; i++) {
    try {
      let response = await axios.post(url, body, {
        headers: { "Content-Type": "text/plain" },
        responseType: "text"
      });
      return response;
    } catch (e) {
      console.log("connection error: " + e.message);
      standardWait();
    }
  }
  throw new Error(
    "connection to the blockchain failed after " + retries + " retries"
  );
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

  // Sends a Message upstream, signing with our keypair.
  // Returns a promise for the response Message.
  // If the response is an error message, we throw an error with the provided error string.
  async sendMessage(message) {
    let clientMessage = SignedMessage.fromSigning(message, this.keyPair);
    let url = getServerURL() + "/messages";
    let body = clientMessage.serialize() + "\n";
    this.log("sending body:", body);
    let response = await retryPost(url, body, {
      headers: { "Content-Type": "text/plain" },
      responseType: "text"
    });
    let text = response.data;
    let serialized = text.replace(/\n$/, "");

    // When there is an empty keepalive message from the server, we just return null
    let signed = SignedMessage.fromSerialized(serialized);
    if (signed == null) {
      return null;
    }

    if (signed.message.type === "Error") {
      console.log("error sending message:", message);
      throw new Error(signed.message.error);
    }
    return signed.message;
  }

  // Sends a query message. Returns the data message response.
  // Throws an error if we get an error message back.
  async query(params) {
    let qm = new Message("Query", params);
    let dm = await this.sendMessage(qm);
    return dm;
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

    // Wait for the op to be processed
    while (true) {
      let account = await this.getAccount(user);
      if (account.sequence === newSequence) {
        break;
      }
      await standardWait();
    }
  }

  // Fetches the account with the given user, or null if there is no such account.
  async getAccount(user) {
    let dm = await this.query({ account: user });
    if (!dm.accounts || !dm.accounts[user]) {
      return null;
    }
    return dm.accounts[user];
  }

  // Fetches data for providers according to the given query.
  // "owner" and "bucket" keys are the most likely.
  // Returns a list of providers in no particular order.
  async getProviders(query) {
    let dm = await this.query({ providers: query });
    return dm.providers;
  }

  // Fetches the provider with the given provider id.
  // If there is no such provider, returns null.
  async getProvider(providerID) {
    let providers = await this.getProviders({ id: providerID });
    if (!providers || providers.length == 0) {
      return null;
    }
    return providers[0];
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

  // Returns the information for the newly-created provider.
  async createProvider(capacity) {
    if (typeof capacity !== "number") {
      throw new Error(
        "capacity " + capacity + " must be number, not " + typeof capacity
      );
    }

    // To figure out which provider is newly-created, we need to check existing ones
    let user = this.keyPair.getPublicKey();
    let initialProviders = await this.getProviders({ owner: user });
    this.log("existing providers:", initialProviders);
    await this.sendOperation("CreateProvider", { capacity });
    this.log("the CreateProvider operation has been accepted");
    let providers = await this.getProviders({ owner: user });
    this.log("new providers:", providers);

    for (let provider of providers) {
      if (!initialProviders.find(p => p.id === provider.id)) {
        // This provider wasn't in the initial set
        return provider;
      }
    }

    throw new Error("no provider seems to have been created");
  }

  // Fetches data for the listed buckets.
  // Returns a list of buckets in no particular order.
  async getBuckets(query) {
    let dm = await this.query({ buckets: query });
    return dm.buckets;
  }

  // Fetches data for the named bucket.
  // Returns null if there is no such bucket.
  async getBucket(name) {
    let buckets = await this.getBuckets({ name });
    if (buckets.length < 1) {
      return null;
    }
    return buckets[0];
  }

  // Returns the information for the newly-created bucket.
  async createBucket(name, size) {
    if (typeof size !== "number") {
      throw new Error(
        "bucket size " + size + " must be number, not " + typeof size
      );
    }

    await this.sendOperation("CreateBucket", { name, size });
    this.log("the CreateBucket operation has been accepted");
    let bucket = await this.getBucket(name);
    return bucket;
  }

  async updateBucket(name, magnet) {
    await this.sendOperation("UpdateBucket", { name, magnet });
    this.log("the UpdateBucket operation has been accepted");
    let bucket = await this.getBucket(name);
    return bucket;
  }

  async allocate(bucketName, providerID) {
    if (typeof providerID !== "number") {
      throw new Error(
        "provider id " +
          providerID +
          " must be number, not " +
          typeof providerID
      );
    }

    await this.sendOperation("Allocate", { bucketName, providerID });
    this.log("the Allocate operation has been accepted");
    return;
  }

  async deallocate(bucketName, providerID) {
    if (typeof providerID !== "number") {
      throw new Error(
        "provider id " +
          providerID +
          " must be number, not " +
          typeof providerID
      );
    }

    await this.sendOperation("Deallocate", { bucketName, providerID });
    this.log("the Deallocate operation has been accepted");
    return;
  }
}

module.exports = ChainClient;