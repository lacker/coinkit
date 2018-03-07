# coinkit
Tools for making cryptocurrency stuff.

## What is this

This code runs a custom blockchain protocol on which you can own and transfer
cryptocurrency.

The consensus mechanism is based on the Stellar Consensus Protocol,
aka SCP. See:

https://www.stellar.org/papers/stellar-consensus-protocol.pdf 

## How to install it

I provide OS X instructions only. Good luck.

First install Postgres and create a test database.

```
brew install postgres
brew services start postgresql
```

Then install go on your machine.

```
brew install go
```

You will need to set up a `GOPATH`, and then clone this repo into the `src`
directory under your gopath. If you have no other preference, I suggest making
`~/go` your `GOPATH` and cloning this repo into `~/go/src/coinkit`.

When you build this repo, it creates multiple binaries in `$GOPATH/bin`.
I suggest adding `$GOPATH/bin` to your `$PATH` - if you don't, you'll have to run
`$GOPATH/bin/cclient` instead of just `cclient`, and so on.

```
# Install dependencies
cd ~/go/src/coinkit
go get -t ./...

# Create databases
./create-databases.sh

# Run the unit tests
go test ./...

# Build everything
go install ./...
```

## How to run it

Commands are from the `~/go/src/coinkit` directory.

To run a local cluster of four cservers:

```
./start-local.sh
```

To stop the local cluster:

```
./stop-local.sh
```

You can check the current account balance with:

```
cclient status [publicKey]
```

If no key is provided, it will prompt you for a passphrase. You can also
use this to create your own account - just use any passphrase, and then
note what the public key is so that other accounts can send you money.

To send money:

```
cclient send [user] [amount]
```

The send command will keep checking back to see when the money leaves the source
account. It should just take a second or two to send the money.

To start off with, all the money is in one account where the passphrase is "mint".
If you're just poking around, I recommend sending some money from the mint
to an account of your own and then checking your account's balance as a little
exercise.

## Benchmarking

```
# Testing a 4-server network with one client
go test ./network -run=zzz -bench=BenchmarkSendMoney1$ -benchtime=20s

# With 10 clients
go test ./network -run=zzz -bench=BenchmarkSendMoney10$ -benchtime=20s

# With 30 clients
go test ./network -run=zzz -bench=BenchmarkSendMoney30$ -benchtime=20s
```

## Code organization

* `cmd`: The code for the command-line tools, `cserver` and `cclient`.
* `consensus`: The logic to run the SCP. This is how blocks are formed.
* `currency`: The financial logic for accounts to process transactions.
* `data`: The code that interacts with Postgres to store past blocks.
* `network`: The networking wrapper to run a server and communicate with peers.
