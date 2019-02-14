let readline = require("readline");

import KeyPair from "./KeyPair";

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

async function generate() {
  let kp = await login();
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
    fatal("usage: npm run cli <operation> <arguments>");
  }

  let op = args[0];
  let rest = args.slice(1);

  if (op === "generate") {
    if (rest.length != 0) {
      fatal("usage: npm run cli generate");
    }

    await generate();
    return;
  }

  fatal("unrecognized operation: " + op);
}

main().then(() => {
  console.log("done");
});
