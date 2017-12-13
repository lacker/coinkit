package main

import (
	"fmt"
	"log"
	"math"
	
	"coinkit/auth"
)

const (
	BasePort = 9000
	NumPeers = 4
)

type Config struct {
	// What port we should listen on
	Port int
	
	// The physical network config of who to send messages to
	PeerPorts []int

	// Our own identity
	KeyPair *auth.KeyPair

	// Defining our quorum
	Members []string
	Threshold int
}

func LocalKeyPair(arg int) *auth.KeyPair {
	return auth.NewKeyPairFromSecretPhrase(fmt.Sprintf("localnet node %d", arg))
}

func NewLocalConfig(arg int) *Config {
	if arg < 0 || arg >= NumPeers {
		log.Fatalf("invalid arg: %d", arg)
	}
	port := BasePort + arg
	kp := LocalKeyPair(arg)

	var peerPorts []int
	var members []string
	for i := 0; i < NumPeers; i++ {
		members = append(members, LocalKeyPair(i).PublicKey())
		p := BasePort + i
		if p == port {
			continue
		}
		peerPorts = append(peerPorts, p)
	}

	// Require a 2k+1 out of 3k+1 consensus
	threshold := int(math.Ceil(2.0 / 3.0 * float64(len(members) - 1))) + 1

	return &Config{
		Port: port,
		KeyPair: kp,
		Members: members,
		Threshold: threshold,
	}
}
