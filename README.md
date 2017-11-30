# coinkit
Tools for making cryptocurrency stuff.

Currently main.go contains a small server that connects to a list of other
nodes and regularly sends some simple stats out to the network.

The next big milestone is for these nodes to operate a blockchain. Let's try the
Stellar Consensus Protocol, aka SCP. See:

https://www.stellar.org/papers/stellar-consensus-protocol.pdf 

Next step: give the nodes a cryptographic identity rather than a port-based one