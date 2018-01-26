# coinkit
Tools for making cryptocurrency stuff.

## What is this

This code runs a custom blockchain protocol on which you can own and transfer
cryptocurrency.

The consensus mechanism is based on the Stellar Consensus Protocol,
aka SCP. See:

https://www.stellar.org/papers/stellar-consensus-protocol.pdf 

## How to install it

To run this, first install go on your machine.

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
# First install dependencies
cd ~/go/src/coinkit
go get -t ./...

# Run the unit tests
go test ./...

# Build everything
go install ./...

## How to run it

TODO: explain both how to run a local cluster and how to connect to a remote one

# Run one of these commands, to run server 0, 1, 2, or 3
cserver 0
cserver 1
cserver 2
cserver 3
```

## Code organization

* `cmd`: The code for the command-line tools, `cserver` and `cclient`.
* `consensus`: The logic to run the SCP. This is how blocks are formed.
* `currency`: The financial logic for accounts to process transactions.
* `network`: The networking wrapper to run a server and communicate with peers.
