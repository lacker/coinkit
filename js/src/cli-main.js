let args = process.argv.slice(2);

if (args.length == 0) {
  console.log("No command line arguments found");
  process.exit(1);
}

args.forEach(val => {
  console.log("CLI arg:", val);
});
