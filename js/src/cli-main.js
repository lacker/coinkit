function fatal(message) {
  console.log(message);
  process.exit(1);
}

function generate() {
  fatal("TODO: implement generate");
}

function main() {
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

    generate();
  }
}

main();
