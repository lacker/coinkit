package network

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"coinkit/consensus"
	"coinkit/util"
)

type Address struct {
	Host string
	Port int
}

func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// Configuration for a network
type NetworkConfig struct {
	// Nodes that are accepting external connections for this network
	Nodes []*Address

	// Defining the quorum for the network
	Members   []string
	Threshold int
}

// Configuration for a particular server running part of the network
type ServerConfig struct {
	Network *NetworkConfig

	Port    int
	KeyPair *util.KeyPair
}

func (nc *NetworkConfig) QuorumSlice() consensus.QuorumSlice {
	return consensus.MakeQuorumSlice(nc.Members, nc.Threshold)
}

// Using a seed prevents multiple networks from accidentally communicating
// with each other if you don't want them to. If you do want different
// programs to communicate with each other on a localhost network, just
// use the same seed, like zero.
func NewLocalhostNetwork(
	firstPort int, num int, seed int) (*NetworkConfig, []*ServerConfig) {

	// Require a 2k+1 out of 3k+1 consensus
	threshold := int(math.Ceil(2.0/3.0*float64(num-1))) + 1

	network := &NetworkConfig{
		Nodes:     []*Address{},
		Members:   []string{},
		Threshold: threshold,
	}
	servers := []*ServerConfig{}

	for port := firstPort; port < firstPort+num; port++ {
		network.Nodes = append(network.Nodes, &Address{
			Host: "127.0.0.1",
			Port: port,
		})
		kp := util.NewKeyPairFromSecretPhrase(fmt.Sprintf("%d %d", seed, port))
		network.Members = append(network.Members, kp.PublicKey().String())
		servers = append(servers, &ServerConfig{
			Network: network,
			Port:    port,
			KeyPair: kp,
		})
	}

	return network, servers
}

const MinUnitTestPort = 2000
const MaxUnitTestPort = 8999

var nextUnitTestPort = MinUnitTestPort

// Avoids port contention which slows down the tests that use ports
func NewUnitTestNetwork() (*NetworkConfig, []*ServerConfig) {
	rand.Seed(int64(time.Now().Nanosecond()))
	num := 4
	if nextUnitTestPort+num > MaxUnitTestPort {
		nextUnitTestPort = MinUnitTestPort
	}
	n, s := NewLocalhostNetwork(nextUnitTestPort, num, rand.Int())
	nextUnitTestPort += num
	return n, s
}

func NewLocalNetwork() (*NetworkConfig, []*ServerConfig) {
	return NewLocalhostNetwork(9000, 4, 0)
}

// Just returns a port
func (nc *NetworkConfig) RandomAddress() *Address {
	rand.Seed(int64(time.Now().Nanosecond()))
	index := rand.Intn(len(nc.Nodes))
	return nc.Nodes[index]
}
