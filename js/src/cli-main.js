const readline = require("readline");

const ChainClient = require("./ChainClient.js");
const KeyPair = require("./KeyPair.js");
const Message = require("./Message.js");
const ProviderListener = require("./ProviderListener.js");

function fatal(message) {
  console.log(message);
  process.exit(1);
}

// Asks the CLI user a question, asynchronously returns the response.
async function ask(question, hideResponse) {
  let r = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });

  let p = new Promise((resolve, reject) => {
    r.question(question, answer => {
      r.close();
      resolve(answer);
    });
    if (hideResponse) {
      r.stdoutMuted = true;
      r._writeToOutput = () => {
        r.output.write("*");
      };
    }
  });

  let answer = await p;
  if (hideResponse) {
    console.log();
  }
  return answer;
}

// Fetches, displays, and returns the account data for a user.
async function status(user) {
  let client = new ChainClient();
  let account = await client.getAccount(user);
  if (!account) {
    console.log("no account found for user", user);
    return null;
  }

  console.log("account data:");
  console.log(account);
  return account;
}

// Asks for a login then displays the status
async function ourStatus() {
  let kp = await login();
  await status(kp.getPublicKey());
}

async function generate() {
  let kp = await login();
  console.log(kp.serialize());
  console.log("key pair generation complete");
}

async function getProvider(id) {
  let client = new ChainClient();
  let provider = await client.getProvider(id);
  if (provider) {
    console.log(provider);
  } else {
    console.log("no provider with id", id);
  }
}

async function getProviders(query) {
  let client = new ChainClient();
  let providers = await client.getProviders(query);
  let word = providers.length === 1 ? "provider" : "providers";
  console.log(providers.length + " " + word + " found");
  for (let p of providers) {
    console.log(p);
  }
}

async function createProvider(capacity) {
  let kp = await login();
  let client = new ChainClient(kp);
  let provider = await client.createProvider(capacity);
  console.log("created provider:");
  console.log(provider);
}

async function getBucket(name) {
  let client = new ChainClient();
  let bucket = await client.getBucket(name);
  if (bucket) {
    console.log(bucket);
  } else {
    console.log("no bucket with name " + name);
  }
}

async function getBuckets(query) {
  let client = new ChainClient();
  let buckets = await client.getBuckets(query);
  let word = buckets.length === 1 ? "bucket" : "buckets";
  console.log(buckets.length + " " + word + " found");
  for (let b of buckets) {
    console.log(b);
  }
}

async function createBucket(name, size) {
  let kp = await login();
  let client = new ChainClient(kp);
  let bucket = await client.createBucket(name, size);
  console.log("created bucket:");
  console.log(bucket);
}

// Ask the user for a passphrase to log in.
// Returns the keypair
async function login() {
  let phrase = await ask("please enter your passphrase: ", true);
  let kp = KeyPair.fromSecretPhrase(phrase);
  console.log("hello. your name is", kp.getPublicKey());
  return kp;
}

async function main() {
  let args = process.argv.slice(2);

  if (args.length == 0) {
    fatal("Usage: npm run cli <operation> <arguments>");
  }

  let op = args[0];
  let rest = args.slice(1);

  if (op === "status") {
    if (rest.length > 1) {
      fatal("Usage: npm run cli status [publickey]");
    }
    if (rest.length === 0) {
      await ourStatus();
    } else {
      await status(rest[0]);
    }
    return;
  }

  if (op === "generate") {
    if (rest.length != 0) {
      fatal("Usage: npm run cli generate");
    }

    await generate();
    return;
  }

  if (op === "create-provider") {
    if (rest.length != 1) {
      fatal("Usage: npm run cli create-provider <capacity>");
    }

    let capacity = parseInt(rest[0]);
    if (!capacity) {
      fatal("bad argument:", rest[0]);
    }
    await createProvider(capacity);
    return;
  }

  if (op === "get-provider") {
    if (rest.length != 1) {
      fatal("Usage: npm run cli get-provider <id>");
    }
    let id = parseInt(rest[0]);
    if (!id) {
      fatal("bad provider id argument:", rest[0]);
    }
    await getProvider(id);
    return;
  }

  if (op === "get-providers") {
    if (rest.length > 2) {
      fatal("Usage: npm run cli get-providers [owner=<id>] [bucket=<name>]");
    }
    let query = {};
    for (let arg of rest) {
      if (arg.startsWith("owner=")) {
        query.owner = arg.split("=")[1];
      } else if (arg.startsWith("bucket=")) {
        query.bucket = arg.split("=")[1];
      } else {
        fatal("unrecognized arg: " + arg);
      }
    }
    if (rest.length === 0) {
      let kp = await login();
      console.log("fetching your providers");
      query.owner = kp.getPublicKey();
    }
    await getProviders(query);
    return;
  }

  if (op === "create-bucket") {
    if (rest.length != 2) {
      fatal("Usage: npm run cli create-bucket <name> <size>");
    }
    let name = rest[0];
    let size = parseInt(rest[1]);
    if (!size) {
      fatal("bad size:", rest[1]);
    }
    await createBucket(name, size);
    return;
  }

  if (op === "get-bucket") {
    if (rest.length != 1) {
      fatal("Usage: npm run cli get-bucket <name>");
    }
    await getBucket(rest[0]);
    return;
  }

  if (op === "get-buckets") {
    if (rest.length > 2) {
      fatal("Usage: npm run cli get-buckets [owner=<id>] [provider=<id>]");
    }
    let query = {};
    for (let arg of rest) {
      if (arg.startsWith("owner=")) {
        query.owner = arg.split("=")[1];
      } else if (arg.startsWith("provider=")) {
        let rhs = arg.split("=");
        let id = parseInt(rhs);
        if (!id) {
          fatal("bad provider id: " + rhs);
        }
        query.provider = id;
      } else {
        fatal("unrecognized arg: " + arg);
      }
    }
    if (rest.length === 0) {
      let kp = await login();
      console.log("fetching your buckets");
      query.owner = kp.getPublicKey();
    }
    await getBuckets(query);
    return;
  }

  if (op === "listen") {
    if (rest.length != 1) {
      fatal("Usage: npm run cli listen [providerID]");
    }

    let id = parseInt(rest[0]);
    if (!id) {
      fatal("bad id: " + rest[0]);
    }
    let listener = new ProviderListener();
    listener.verbose = true;
    await listener.listen(id);
    return;
  }

  fatal("unrecognized operation: " + op);
}

main()
  .then(() => {
    // console.log("done");
  })
  .catch(e => {
    fatal("Unhandled " + e);
  });
