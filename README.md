# coinkit
Tools for making cryptocurrency stuff.

This runs a server that creates a blockchain using the Stellar Consensus Protocol, aka SCP. See:

https://www.stellar.org/papers/stellar-consensus-protocol.pdf 

The current configuration is set to run a 4-node network on your local machine.
The network should be resistant to 1 of the 4 nodes being offline or malicious.

To run this, first install go on your machine.

```
brew install go
```

You will need to set up a `GOPATH`, and then clone this repo into the `src` directory
under your gopath. If you have no other preference, I suggest making `~/go` your
`GOPATH` and cloning this repo into `~/go/src/coinkit`.

Then, the simplest way to run some local servers is to run them in four separate
terminals.

```
# First install dependencies
cd ~/go/src/coinkit
go get -t ./...

# If you would like to run the unit tests
go test ./...

# And then run one of these commands, to run server 0, 1, 2, or 3
go run main.go 0
go run main.go 1
go run main.go 2
go run main.go 3
```

You can run just three out of the four if you so desire.

The nodes will start at the genesis block and proceed through the blockchain.
When it says "advancing to slot X" that indicates it has finished consensus on
a block and is moving to the next one. Other log messages indicate various
phases of this process.

TODO: add more features

There is an artificial 1-2 second pause between messages, set in `server.go`, to
make it simpler to see what is going on.