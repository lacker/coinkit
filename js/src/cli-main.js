let readline = require("readline");

import ChainClient from "./ChainClient";
import KeyPair from "./KeyPair";
import Message from "./Message";

function fatal(message) {
  console.log(message);
  process.exit(1);
}

// Asks the CLI user a question, asynchronously returns the response
async function ask(question) {
  let r = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });

  let p = new Promise((resolve, reject) => {
    r.question(question, answer => {
      r.close();
      resolve(answer);
    });
  });

  return await p;
}

// Fetches, displays, and returns the account data for a user.
async function status(user) {
  let client = new ChainClient();
  let qm = new Message({
    account: user
  });
  let dm = await client.sendMessage(qm);

  if (!dm.accounts || !dm.accounts[user]) {
    console.log("no account found for user", user);
    return;
  }
  let account = dm.accounts[user];

  console.log("account data for", user + ":", account);
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

// Ask the user for a passphrase to log in.
// Returns the keypair
async function login() {
  let phrase = await ask("please enter your passphrase:");
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
      ourStatus();
    } else {
      status(rest[0]);
    }
  }

  if (op === "generate") {
    if (rest.length != 0) {
      fatal("Usage: npm run cli generate");
    }

    await generate();
    return;
  }

  fatal("unrecognized operation: " + op);
}

main().then(() => {
  // console.log("done");
});
