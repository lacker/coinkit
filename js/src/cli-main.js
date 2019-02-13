let args = process.argv.slice(2);

function fatal(message) {
  console.log(message);
  process.exit(1);
}

if (args.length == 0) {
  fatal("usage: npm run cli <operation> <arguments>");
}

let op = args[0];
let rest = args.slice(1);

if (op === "generate") {
  if (rest.length != 0) {
    fatal("usage: npm run cli generate");
  }

  generate();
}

function generate() {
  fatal("TODO: implement generate");
}
